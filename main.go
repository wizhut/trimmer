package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const version = "1.0.0"

// companyNotice is appended to the usage template, so it shows up under the
// flags on every help path — `--help`, `--version`, and the usage text printed
// when trimmer is run with no piped stdin.
const companyNotice = `Provided by wizhut.tech (https://wizhut.tech).
Free to use for any purpose. Issues and contributions welcome.`

var (
	perLine     bool
	leftOnly    bool
	rightOnly   bool
	noNewline   bool
	showVersion bool
)

var rootCmd = &cobra.Command{
	Use:   "trimmer [flags]",
	Short: "Strips leading and trailing whitespace from stdin",
	Long: `Reads stdin and writes it back to stdout with the leading and trailing
whitespace removed.

By default the whole input is treated as a single value, so only the very start
and the very end are trimmed. Pass --lines to trim each line on its own, and
--left / --right to trim one side only.

Whitespace is Unicode-aware: spaces, tabs, newlines, carriage returns and any
other rune matching unicode.IsSpace.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Fprintf(cmd.OutOrStdout(), "trimmer version %s\n\n%s\n", version, companyNotice)
			return nil
		}

		if isTerminal(os.Stdin) {
			return cmd.Usage()
		}

		return run(os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + "\n" + companyNotice + "\n")

	rootCmd.Flags().BoolVarP(&perLine, "lines", "l", false, "Trim each line individually instead of the whole input")
	rootCmd.Flags().BoolVar(&leftOnly, "left", false, "Trim leading whitespace only")
	rootCmd.Flags().BoolVar(&rightOnly, "right", false, "Trim trailing whitespace only")
	rootCmd.Flags().BoolVarP(&noNewline, "no-newline", "n", false, "Do not terminate the output with a newline")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show version information")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "trimmer:", err)
		os.Exit(1)
	}
}

// run applies the current flag configuration to r and writes the result to w.
func run(r io.Reader, w io.Writer) error {
	out := bufio.NewWriter(w)
	defer out.Flush()

	if perLine {
		return trimLines(r, out)
	}
	return trimAll(r, out)
}

// trimAll treats the entire input as one value: only its outermost whitespace
// goes away. It has to buffer everything, since the trailing run of whitespace
// is only known once the input ends.
func trimAll(r io.Reader, w *bufio.Writer) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := w.WriteString(trim(string(b))); err != nil {
		return err
	}
	if !noNewline {
		return w.WriteByte('\n')
	}
	return nil
}

// trimLines streams the input a line at a time, trimming each one. Line count is
// preserved — blank lines stay blank rather than disappearing — and an input
// that ends without a newline produces output that does too.
func trimLines(r io.Reader, w *bufio.Writer) error {
	br := bufio.NewReader(r)

	// The terminator of a line is written lazily, when the next line shows up.
	// That way the final one can still be dropped by --no-newline.
	pending := false

	for {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if line == "" && err == io.EOF {
			break
		}

		if pending {
			if werr := w.WriteByte('\n'); werr != nil {
				return werr
			}
		}
		if _, werr := w.WriteString(trim(strings.TrimSuffix(line, "\n"))); werr != nil {
			return werr
		}
		pending = strings.HasSuffix(line, "\n")

		if err == io.EOF {
			break
		}
	}

	if pending && !noNewline {
		return w.WriteByte('\n')
	}
	return nil
}

// trim strips whitespace from the side(s) selected by --left / --right. Passing
// neither (or both) trims both sides.
func trim(s string) string {
	switch {
	case leftOnly && !rightOnly:
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	case rightOnly && !leftOnly:
		return strings.TrimRightFunc(s, unicode.IsSpace)
	default:
		return strings.TrimSpace(s)
	}
}

// isTerminal reports whether f is an interactive terminal, i.e. nothing was
// piped or redirected in. A plain ModeCharDevice check is not enough here:
// /dev/null is a character device too, and `trimmer < /dev/null` is a real
// (if pointless) invocation that should produce output, not usage text.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}
