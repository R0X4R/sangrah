package cmd

import (
	"fmt"

	"github.com/projectdiscovery/goflags"
)

// Options holds all user-configurable flags for sangrah.
type Options struct {
	InputFile string
	Headers   goflags.StringSlice
	Cookie    string
	Threads   int
	Retries   int
	OutputDir string
	Silent    bool
}

// ParseOptions reads CLI flags using goflags and returns the parsed Options.
// Prints help and exits on --help; returns an error on invalid flags.
func ParseOptions() (*Options, error) {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`Sangrah - Download, format, and save JavaScript files from a list of URLs.`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.InputFile, "input", "i", "", "\tFile containing URLs (one per line, or pipe to stdin)"),
	)

	flagSet.CreateGroup("request", "Request",
		flagSet.StringSliceVarP(&options.Headers, "header", "H", nil, "\tCustom header (repeatable, e.g. -H 'Authorization: Bearer xxx')", goflags.StringSliceOptions),
		flagSet.StringVarP(&options.Cookie, "cookie", "c", "", "\tCookie header value"),
	)

	flagSet.CreateGroup("runtime", "Performance",
		flagSet.IntVarP(&options.Threads, "threads", "t", 10, "\tNumber of concurrent downloads"),
		flagSet.IntVarP(&options.Retries, "retries", "r", 3, "\tRetry count on failure"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.OutputDir, "output", "o", ".", "\tOutput directory for downloaded JS files"),
		flagSet.BoolVarP(&options.Silent, "silent", "s", false, "\tOnly show [FAIL] output"),
	)

	if err := flagSet.Parse(); err != nil {
		return nil, err
	}

	if options.Threads <= 0 {
		return nil, fmt.Errorf("threads must be greater than 0")
	}
	if options.Retries < 0 {
		return nil, fmt.Errorf("retries must not be negative")
	}

	return options, nil
}
