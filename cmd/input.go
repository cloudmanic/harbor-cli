// Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
// Date: 2026-06-22

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// defaultContentFormat is the CLI's default interpretation of body content.
// Markdown is the friendliest input for humans and agents; the server converts
// it to sanitized HTML. Override per command with --format html.
const defaultContentFormat = "markdown"

// addContentFlags registers the shared body-input flags (--content / --file /
// --stdin / --format) on a command that creates or updates note-like bodies.
func addContentFlags(cmd *cobra.Command) {
	cmd.Flags().String("content", "", "Body content as a literal string")
	cmd.Flags().String("file", "", "Read body content from a file")
	cmd.Flags().Bool("stdin", false, "Read body content from standard input")
	cmd.Flags().String("format", defaultContentFormat, "Content format: markdown or html")
}

// readContent resolves body content from the mutually-exclusive --content /
// --file / --stdin flags. It returns the content, the chosen format, and a
// hasContent flag (false when none of the three was provided, which is valid
// for updates that change only metadata). At most one source may be set.
func readContent(cmd *cobra.Command) (content, format string, hasContent bool, err error) {
	format = stringFlag(cmd, "format")
	if format == "" {
		format = defaultContentFormat
	}
	if format != "markdown" && format != "html" {
		return "", "", false, fmt.Errorf("invalid --format %q (want markdown or html)", format)
	}

	literal := cmd.Flags().Changed("content")
	file := stringFlag(cmd, "file")
	stdin := boolFlag(cmd, "stdin")

	sources := 0
	if literal {
		sources++
	}
	if file != "" {
		sources++
	}
	if stdin {
		sources++
	}
	if sources > 1 {
		return "", "", false, errors.New("use only one of --content, --file, or --stdin")
	}
	if sources == 0 {
		return "", format, false, nil
	}

	switch {
	case literal:
		content = stringFlag(cmd, "content")
	case file != "":
		b, rerr := os.ReadFile(file)
		if rerr != nil {
			return "", "", false, fmt.Errorf("failed to read --file: %w", rerr)
		}
		content = string(b)
	case stdin:
		b, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return "", "", false, fmt.Errorf("failed to read stdin: %w", rerr)
		}
		content = string(b)
	}
	return content, format, true, nil
}
