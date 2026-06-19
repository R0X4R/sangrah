// Sangrah downloads JavaScript files from URLs, beautifies them with
// 4-space indentation, and saves them with sanitized filenames.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/R0X4R/sangrah/cmd"
	"github.com/fatih/color"
)

// nonAlphaNum matches any character that isn't a letter, digit, or hyphen.
var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9-]`)

var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

func main() {
	opts, err := cmd.ParseOptions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	urls := readURLs(opts)
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: no URLs provided. Pipe to stdin or use -i <file>")
		os.Exit(1)
	}

	cmd.InitClient(opts.Threads)

	sem := make(chan struct{}, opts.Threads)
	var wg sync.WaitGroup
	var failed atomic.Int64

	for _, u := range urls {
		wg.Add(1)
		go func(rawURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			data, fetchErr := cmd.FetchJS(rawURL, opts)
			if fetchErr != nil {
				fmt.Fprintf(os.Stderr, "%s %s: %v\n", red("[FAIL]"), rawURL, fetchErr)
				failed.Add(1)
				return
			}

			beautified := cmd.BeautifyJS(data)
			filename := sanitizeFilename(rawURL)
			outputPath := filepath.Join(opts.OutputDir, filename)

			os.MkdirAll(opts.OutputDir, 0755)
			if writeErr := os.WriteFile(outputPath, beautified, 0644); writeErr != nil {
				fmt.Fprintf(os.Stderr, "%s %s: %v\n", red("[FAIL]"), rawURL, writeErr)
				failed.Add(1)
				return
			}

			if !opts.Silent {
				fmt.Fprintf(os.Stdout, "%s %s\n", green("[SAVED]"), rawURL)
			}
		}(u)
	}

	wg.Wait()

	if failed.Load() > 0 {
		os.Exit(1)
	}
}

// readURLs collects URLs from either a file (-i flag) or stdin.
// Lines starting with # are treated as comments and skipped.
func readURLs(opts *cmd.Options) []string {
	var urls []string

	if opts.InputFile != "" {
		data, err := os.ReadFile(opts.InputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading %s: %v\n", opts.InputFile, err)
			os.Exit(1)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
		return urls
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	return urls
}

// sanitizeFilename converts a URL into a safe filename by replacing any
// character that isn't a letter, digit, or hyphen with underscores.
// The trailing .js is preserved without duplication.
func sanitizeFilename(rawURL string) string {
	rest := rawURL
	if strings.HasSuffix(strings.ToLower(rest), ".js") {
		rest = rest[:len(rest)-3]
	}
	return nonAlphaNum.ReplaceAllString(rest, "_") + ".js"
}
