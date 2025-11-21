package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

// urlEntry holds the original URL string and its sort key components
type urlEntry struct {
	original string
	sortKey  sortKey
}

// sortKey contains all components used for sorting
type sortKey struct {
	domain   string // reversed domain components (case-insensitive comparison)
	port     int    // numeric port value
	scheme   string // scheme (case-insensitive comparison)
	path     string // path (case-sensitive)
	query    string // query string (case-sensitive)
	fragment string // fragment (case-sensitive)
}

// schemeDefaultPorts maps common schemes to their default ports
var schemeDefaultPorts = map[string]int{
	"http":  80,
	"https": 443,
	"ftp":   21,
	"ssh":   22,
	"ws":    80,
	"wss":   443,
	"file":  -1, // -1 means no port
}

func help() {
	fmt.Fprint(os.Stderr, ""+
		"urlsort - sorts URLs based on the  components of the url.\n\n"+

		"Usage: urlsort [OPTIONS] [FILE...]\n\n"+

		"Reads URLs from standard input or specified files, sorts them,\n"+
		"and writes the output.\n\n"+

		"sorts by domain, port, scheme, path, querystring, then fragment\n\n"+

		"Options:\n",
	)
	pflag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	var outputFile string
	var helpFlag bool
	pflag.StringVarP(&outputFile, "output-file", "o", "", "write output to file")
	pflag.BoolVarP(&helpFlag, "help", "h", false, "this help output")
	pflag.Parse()

	if helpFlag {
		help()
		os.Exit(0)
	}

	// Collect all input sources
	var urls []string
	args := pflag.Args()

	if len(args) == 0 {
		// Read from stdin
		urls = readFromReader(os.Stdin)
	} else {
		// Read from files and stdin (if - is specified)
		for _, arg := range args {
			if arg == "-" {
				urls = append(urls, readFromReader(os.Stdin)...)
			} else {
				fileURLs, err := readFromFile(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", arg, err)
					os.Exit(1)
				}
				urls = append(urls, fileURLs...)
			}
		}
	}

	// Parse and create sortable entries
	entries := make([]urlEntry, 0, len(urls))
	for _, urlStr := range urls {
		entry := parseURL(urlStr)
		entries = append(entries, entry)
	}

	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		return compareSortKeys(entries[i].sortKey, entries[j].sortKey)
	})

	// Determine output destination
	var writer io.Writer
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		writer = file
	} else {
		writer = os.Stdout
	}

	// Write sorted URLs
	for _, entry := range entries {
		fmt.Fprintln(writer, entry.original)
	}
}

// readFromReader reads URLs from an io.Reader, one per line
func readFromReader(reader io.Reader) []string {
	var urls []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		urls = append(urls, line)
	}
	return urls
}

// readFromFile reads URLs from a file, one per line
func readFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readFromReader(file), nil
}

// parseURL parses a URL string and extracts sort key components
func parseURL(urlStr string) urlEntry {
	entry := urlEntry{
		original: urlStr,
		sortKey: sortKey{
			port: -1, // -1 means no port specified
		},
	}

	// Handle empty lines as invalid URLs
	if strings.TrimSpace(urlStr) == "" {
		return entry
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		// Invalid URL - return entry with empty sort key components
		return entry
	}

	// Extract scheme (case-insensitive for comparison, but store lowercase)
	entry.sortKey.scheme = strings.ToLower(parsed.Scheme)

	// Extract and process domain
	host := parsed.Hostname()
	if host != "" {
		entry.sortKey.domain = reverseDomain(host)
	}

	// Extract and process port
	portStr := parsed.Port()
	if portStr != "" {
		port, err := resolvePort(portStr, parsed.Scheme)
		if err == nil {
			entry.sortKey.port = port
		} else {
			// Invalid port, use scheme default
			entry.sortKey.port = getDefaultPort(parsed.Scheme)
		}
	} else {
		// No port specified, use scheme default
		entry.sortKey.port = getDefaultPort(parsed.Scheme)
	}

	// Extract path
	entry.sortKey.path = parsed.Path

	// Extract query string
	entry.sortKey.query = parsed.RawQuery

	// Extract fragment
	entry.sortKey.fragment = parsed.Fragment

	return entry
}

// reverseDomain reverses all domain components for sorting
// IP addresses are kept as-is
func reverseDomain(host string) string {
	// Check if it's an IP address (IPv4 or IPv6)
	if net.ParseIP(host) != nil {
		return strings.ToLower(host)
	}

	// Check if it's an IPv6 address in brackets
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		ipStr := host[1 : len(host)-1]
		if net.ParseIP(ipStr) != nil {
			return strings.ToLower(host)
		}
	}

	// Split domain into components and reverse
	parts := strings.Split(strings.ToLower(host), ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

// resolvePort resolves a port string to a numeric value
// It handles both numeric ports and service names
func resolvePort(portStr, scheme string) (int, error) {
	// Try parsing as number first
	if port, err := strconv.Atoi(portStr); err == nil {
		return port, nil
	}

	// Try looking up as service name
	port, err := net.LookupPort("tcp", portStr)
	if err == nil {
		return port, nil
	}

	// Try UDP if TCP failed
	port, err = net.LookupPort("udp", portStr)
	if err == nil {
		return port, nil
	}

	// If lookup fails, return error
	return 0, fmt.Errorf("cannot resolve port: %s", portStr)
}

// getDefaultPort returns the default port for a scheme
// Returns -1 if the scheme has no port (like file://)
func getDefaultPort(scheme string) int {
	if scheme == "" {
		return -1
	}
	port, ok := schemeDefaultPorts[strings.ToLower(scheme)]
	if ok {
		return port
	}

	// Try to lookup default port for unknown schemes
	// Common schemes: ws, wss, etc.
	lowerScheme := strings.ToLower(scheme)
	if port, ok := schemeDefaultPorts[lowerScheme]; ok {
		return port
	}

	// For unknown schemes, try to lookup standard port
	// This is a best-effort approach
	port, err := net.LookupPort("tcp", scheme)
	if err == nil {
		return port
	}

	// Default to -1 (no port) for unknown schemes
	return -1
}

// compareSortKeys compares two sort keys according to the sorting criteria
func compareSortKeys(a, b sortKey) bool {
	// 1. Domain (case-insensitive)
	if a.domain != b.domain {
		return a.domain < b.domain
	}

	// 2. Port (numeric comparison)
	if a.port != b.port {
		// Handle -1 (no port) - it should sort before any numeric port
		if a.port == -1 {
			return true
		}
		if b.port == -1 {
			return false
		}
		return a.port < b.port
	}

	// 3. Scheme (case-insensitive)
	if a.scheme != b.scheme {
		return a.scheme < b.scheme
	}

	// 4. Path (case-sensitive)
	if a.path != b.path {
		return a.path < b.path
	}

	// 5. Query String (case-sensitive)
	if a.query != b.query {
		return a.query < b.query
	}

	// 6. Fragment (case-sensitive)
	return a.fragment < b.fragment
}
