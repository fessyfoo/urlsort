package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	binPath   string
)

// ensureBinary builds the urlsort binary if it doesn't exist
func ensureBinary(t *testing.T) {
	buildOnce.Do(func() {
		binPath = "./urlsort"
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			cmd := exec.Command("go", "build", "-o", binPath)
			if err := cmd.Run(); err != nil {
				t.Fatalf("failed to build urlsort binary: %v", err)
			}
		}
	})
}

// runURLSort runs the urlsort command with given args and input, returns output and error
func runURLSort(t *testing.T, args []string, input string) (string, string, error) {
	ensureBinary(t)
	cmd := exec.Command(binPath, args...)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output), string(exitErr.Stderr), err
		}
		return string(output), "", err
	}
	return string(output), "", nil
}

func TestDomainSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "reverse domain components",
			input:    "https://www.yahoo.com\nhttps://www.example.com\nhttp://test.com",
			expected: "https://www.example.com\nhttp://test.com\nhttps://www.yahoo.com\n",
		},
		{
			name:     "case insensitive domain",
			input:    "https://EXAMPLE.COM\nhttps://example.com\nhttps://Example.Com",
			expected: "https://EXAMPLE.COM\nhttps://example.com\nhttps://Example.Com\n",
		},
		{
			name:     "subdomain ordering",
			input:    "https://a.example.com\nhttps://z.example.com\nhttps://example.com",
			expected: "https://example.com\nhttps://a.example.com\nhttps://z.example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestIPAddresses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 addresses kept as-is",
			input:    "http://192.168.1.1\nhttp://10.0.0.1\nhttp://172.16.0.1",
			expected: "http://10.0.0.1\nhttp://172.16.0.1\nhttp://192.168.1.1\n",
		},
		{
			name:     "IPv6 addresses kept as-is",
			input:    "http://[2001:db8::2]\nhttp://[2001:db8::1]\nhttp://[::1]",
			expected: "http://[2001:db8::1]\nhttp://[2001:db8::2]\nhttp://[::1]\n",
		},
		{
			name:     "mixed IPs and domains",
			input:    "http://example.com\nhttp://192.168.1.1\nhttp://test.com",
			expected: "http://192.168.1.1\nhttp://example.com\nhttp://test.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestPortSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "numeric port sorting",
			input:    "https://example.com:8080\nhttps://example.com:443\nhttp://example.com:80",
			expected: "http://example.com:80\nhttps://example.com:443\nhttps://example.com:8080\n",
		},
		{
			name:     "default ports applied",
			input:    "https://example.com\nhttp://example.com\nftp://example.com",
			expected: "ftp://example.com\nhttp://example.com\nhttps://example.com\n",
		},
		{
			name:     "explicit vs default ports",
			input:    "http://example.com:80\nhttp://example.com\nhttps://example.com:443",
			expected: "http://example.com:80\nhttp://example.com\nhttps://example.com:443\n",
		},
		{
			name:     "file scheme has no port",
			input:    "file:///path/to/file\nhttp://example.com:80\nhttps://example.com",
			expected: "file:///path/to/file\nhttp://example.com:80\nhttps://example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestSchemeSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "case insensitive scheme",
			input:    "HTTPS://example.com\nhttp://example.com\nHttp://example.com",
			expected: "http://example.com\nHttp://example.com\nHTTPS://example.com\n",
		},
		{
			name:     "different schemes",
			input:    "https://example.com\nftp://example.com\nhttp://example.com",
			expected: "ftp://example.com\nhttp://example.com\nhttps://example.com\n",
		},
		{
			name:     "no scheme sorts as empty",
			input:    "example.com/path\nhttp://example.com/path",
			expected: "example.com/path\nhttp://example.com/path\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestPathSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "case sensitive paths",
			input:    "http://example.com/Path\nhttp://example.com/path\nhttp://example.com/PATH",
			expected: "http://example.com/PATH\nhttp://example.com/Path\nhttp://example.com/path\n",
		},
		{
			name:     "path ordering",
			input:    "http://example.com/z\nhttp://example.com/a\nhttp://example.com/m",
			expected: "http://example.com/a\nhttp://example.com/m\nhttp://example.com/z\n",
		},
		{
			name:     "no path vs empty path",
			input:    "http://example.com/\nhttp://example.com",
			expected: "http://example.com\nhttp://example.com/\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestQueryStringSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "query string comparison",
			input:    "http://example.com?b=2&a=1\nhttp://example.com?a=1&b=2",
			expected: "http://example.com?a=1&b=2\nhttp://example.com?b=2&a=1\n",
		},
		{
			name:     "case sensitive query",
			input:    "http://example.com?A=1\nhttp://example.com?a=1",
			expected: "http://example.com?A=1\nhttp://example.com?a=1\n",
		},
		{
			name:     "no query vs empty query",
			input:    "http://example.com?\nhttp://example.com",
			expected: "http://example.com?\nhttp://example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestFragmentSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fragment comparison",
			input:    "http://example.com#b\nhttp://example.com#a",
			expected: "http://example.com#a\nhttp://example.com#b\n",
		},
		{
			name:     "case sensitive fragment",
			input:    "http://example.com#A\nhttp://example.com#a",
			expected: "http://example.com#A\nhttp://example.com#a\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestComplexSorting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "full sorting criteria",
			input: `https://www.example.com:443/path?q=1#frag
http://test.com:80/path
https://www.example.com:443/path?q=2#frag
http://test.com:80/other`,
			expected: `https://www.example.com:443/path?q=1#frag
https://www.example.com:443/path?q=2#frag
http://test.com:80/other
http://test.com:80/path
`,
		},
		{
			name: "multiple domains and ports",
			input: `https://z.com:8080
http://a.com:80
https://a.com:443
http://z.com:80`,
			expected: `http://a.com:80
https://a.com:443
http://z.com:80
https://z.com:8080
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestInvalidURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "invalid URLs sorted as empty",
			input:    "not a url\nhttp://valid.com\nalso invalid",
			expected: "also invalid\nnot a url\nhttp://valid.com\n",
		},
		{
			name:     "empty lines treated as invalid",
			input:    "http://example.com\n\nhttp://test.com",
			expected: "\nhttp://example.com\nhttp://test.com\n",
		},
		{
			name:     "malformed URLs",
			input:    "://invalid\nhttp://valid.com\nhttp://",
			expected: "://invalid\nhttp://\nhttp://valid.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, nil, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestFileInput(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err := os.WriteFile(file1, []byte("https://z.com\nhttp://a.com"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err = os.WriteFile(file2, []byte("https://m.com\nhttp://b.com"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		input    string
		expected string
	}{
		{
			name:     "single file input",
			args:     []string{file1},
			expected: "http://a.com\nhttps://z.com\n",
		},
		{
			name:     "multiple files merged",
			args:     []string{file1, file2},
			expected: "http://a.com\nhttp://b.com\nhttps://m.com\nhttps://z.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _, err := runURLSort(t, tt.args, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

func TestOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.txt")

	input := "https://z.com\nhttp://a.com"
	expected := "http://a.com\nhttps://z.com\n"

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "short output flag",
			args: []string{"-o", outputFile},
		},
		{
			name: "long output flag",
			args: []string{"--output-file", outputFile},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Remove output file if it exists
			os.Remove(outputFile)

			_, _, err := runURLSort(t, tt.args, input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			if string(content) != expected {
				t.Errorf("expected:\n%s\ngot:\n%s", expected, string(content))
			}
		})
	}
}

func TestStdinWithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")

	err := os.WriteFile(file1, []byte("https://z.com\n"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test stdin in middle of file args
	input := "http://a.com\n"
	args := []string{file1, "-", file1}

	output, _, err := runURLSort(t, args, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "http://a.com\nhttps://z.com\nhttps://z.com\n"
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}

func TestEmptyInput(t *testing.T) {
	output, _, err := runURLSort(t, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

func TestSingleURL(t *testing.T) {
	input := "https://example.com\n"
	output, _, err := runURLSort(t, nil, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != input {
		t.Errorf("expected: %q, got: %q", input, output)
	}
}
