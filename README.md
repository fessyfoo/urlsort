# urlsort

```bash
cat urls.txt | urlssort 
urlsort urls.txt ... 
cat moreurls.txt | urlsort - urls.txt
urlsort urls.txt -o sorted.txt 
```

sorts URLs using a multi-level comparison based on the  components of the url

## Installation

Build the program:

```bash
go build -o urlsort
```

Or install it:

```bash
go install
```

## Usage

### Basic Usage

Read from standard input:

```bash

cat <<EOM | urlsort
http://test.com:81
http://test.com:79
http://test.com/
https://c.example.com/path
https://b.example.com/path
http://b.example.com/path
https://a.example.com/path/b
https://a.example.com/path/a
EOM
```

Read from a file:

```bash
urlsort urls.txt
```

Read from multiple files and stdin:

```bash
urlsort file1.txt - file2.txt
```

The `-` argument reads from stdin at that position in the argument list.

### Output Options

Write output to a file instead of stdout:

```bash
urlsort -o sorted.txt urls.txt
urlsort --output-file sorted.txt urls.txt
```

## Sorting Algorithm

The program sorts URLs using a multi-level comparison based on the following components, in order:

1. **Domain** (case-insensitive)
   - Domain components are reversed for sorting (e.g., `www.yahoo.com` sorts as `com.yahoo.www`)
   - IP addresses (IPv4 and IPv6) are kept as-is and not reversed
   - URLs without domains sort as empty domain

2. **Port** (numeric)
   - Ports are sorted numerically (e.g., 80 < 443 < 8080)
   - Service names (e.g., `:http`, `:https`) are resolved to their numeric values via system service lookup
   - Default ports are used when not specified:
     - `http` → `:80`
     - `https` → `:443`
     - `ftp` → `:21`
     - `ssh` → `:22`
     - `file` → no port
     - Other schemes use their standard default ports (e.g., `ws` → `:80`, `wss` → `:443`)

3. **Scheme** (case-insensitive)
   - Examples: `http`, `https`, `ftp`
   - URLs without schemes sort as having an empty scheme

4. **Path** (case-sensitive)
   - The URL path component

5. **Query String** (case-sensitive)
   - The entire query string is compared lexicographically. *this may change*
   - Original format is preserved (no normalization)

6. **Fragment** (case-sensitive)
   - The URL fragment component

### Error Handling

- Invalid URLs are handled gracefully and sorted as if missing components (empty values for missing parts)
- Empty lines are treated as invalid URLs
- Processing continues even if some URLs are invalid
- Original URL format is preserved exactly in the output (no normalization)

## Examples

```bash
# Sort URLs from stdin
cat urls.txt | urlsort

# Sort URLs from a file and save to another file
urlsort -o sorted.txt urls.txt

# Combine multiple sources and - for stdin
urlsort file1.txt - file2.txt > combined_sorted.txt
```

