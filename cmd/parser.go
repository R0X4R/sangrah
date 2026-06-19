// Package cmd implements the core operations for sangrah — URL fetching,
// JavaScript beautification, and CLI option parsing.
package cmd

import (
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
)

// BeautifyJS pretty-prints JavaScript source with 4-space indentation.
// Falls back to the original input if beautification fails.
func BeautifyJS(input []byte) []byte {
	str := string(input)
	opts := jsbeautifier.DefaultOptions()
	opts["indent_size"] = 4
	opts["indent_char"] = " "

	result, err := jsbeautifier.Beautify(&str, opts)
	if err != nil {
		return input
	}

	return []byte(result)
}
