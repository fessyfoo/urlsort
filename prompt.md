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

1. **Domain** (case-insensitive): Reverse domain components for sorting
   - Example: `www.yahoo.com` sorts as `com.yahoo.www`
   - URLs without domains sort as empty domain

2. **Port** (case-sensitive): After domain in sort order
   - Use scheme-specific default ports if not specified:
     - `http` → `:80`
     - `https` → `:443`
     - `ftp` → `:21`
     - `ssh` → `:22`
     - `file` → no port (empty)
     - Other schemes: use their standard default ports

3. **Scheme** (case-insensitive): e.g., `http`, `https`, `ftp`

4. **Path** (case-sensitive): The URL path component

5. **Query Args** (case-sensitive): Parse query string and sort individual parameters
   - Example: `?b=2&a=1` should be sorted as `?a=1&b=2`
   - Sort parameter names, then values if names are equal

6. **Fragment** (case-sensitive): The URL fragment component

## Error Handling

- Handle invalid URLs gracefully and silently
- For invalid URLs, sort as if missing components (empty values for missing parts)
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
