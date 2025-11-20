# URL Sorter Program Specification

Create a Go program that sorts URLs from standard input or files, similar to the Unix `sort` command.

## Input Handling

- Accept positional arguments (filenames) like the `sort` command
- If positional arguments are provided, read from those files (not stdin)
- If `-` is provided as a positional argument, read from stdin at that position
- Process multiple files/stdin inputs and merge them before sorting
- If no positional arguments are provided, read from stdin
- Expect one URL per line

## Output

- Print sorted URLs to stdout, one per line
- Support GNU-style long options: `-o` or `--output-file` flag to write to a file instead of stdout
- Preserve the original URL format exactly as input (no normalization)

## Sorting Criteria (in order)

1. **Domain** (case-insensitive): Reverse all domain components for sorting
   - Example: `www.yahoo.com` sorts as `com.yahoo.www` (all components reversed)
   - IP addresses (IPv4 and IPv6) are kept as-is (not reversed)
   - URLs without domains sort as empty domain

2. **Port** (case-sensitive): After domain in sort order
   - Sort ports numerically (e.g., 80 < 443 < 8080)
   - Non-numeric port names (e.g., `:http`, `:https`) should be resolved to their numeric values via service name lookup and sorted as numbers
   - Use scheme-specific default ports if not specified:
     - `http` → `:80`
     - `https` → `:443`
     - `ftp` → `:21`
     - `ssh` → `:22`
     - `file` → no port (empty)
     - Other schemes: lookup and use their standard default ports (e.g., `ws` → `:80`, `wss` → `:443`)

3. **Scheme** (case-insensitive): e.g., `http`, `https`, `ftp`
   - URLs without schemes sort as having an empty scheme

4. **Path** (case-sensitive): The URL path component

5. **Query String** (case-sensitive): Compare the query string as a whole
   - Compare the entire query string component lexicographically
   - Preserve original query string format in output (no normalization)

6. **Fragment** (case-sensitive): The URL fragment component

## Error Handling

- Handle invalid URLs gracefully and silently
- For invalid URLs, sort as if missing components (empty values for missing parts)
- Empty lines are treated as invalid URLs and sorted accordingly
- Continue processing even if some URLs are invalid

## Requirements

- Write clean, idiomatic Go code
- Use standard library packages where possible
- Include proper error handling
- Make the code maintainable and well-structured
- Follow Go naming conventions and best practices

## Example Usage

```bash
# Read from stdin
echo -e "https://www.example.com/path\nhttp://test.com" | urlsort

# Read from file
urlsort urls.txt

# Read from multiple files and stdin
urlsort file1.txt - file2.txt

# Output to file
urlsort -o sorted.txt urls.txt
urlsort --output-file sorted.txt urls.txt
```
