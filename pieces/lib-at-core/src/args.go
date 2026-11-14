// Package atcore holds the shared plumbing of the acceptance-test libraries: flags, a registry of
// the repository's JSON Schemas read at runtime, validation, and report emission.
package atcore

import (
	"flag"
	"fmt"
	"io"
	"regexp"
)

// Args holds the command-line flags. Flags are not a schema, so a plain struct is fine.
type Args struct {
	Root   string
	Base   string
	Commit string
}

var commitRe = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Parse reads --root (default "."), --base (optional revision) and --commit (optional, 40 lowercase hex).
func Parse(args []string) (Args, error) {
	var a Args
	fs := flag.NewFlagSet("commons-at", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&a.Root, "root", ".", "repository root")
	fs.StringVar(&a.Base, "base", "", "base revision")
	fs.StringVar(&a.Commit, "commit", "", "commit id, 40 lowercase hex")
	if err := fs.Parse(args); err != nil {
		return Args{}, err
	}
	if fs.NArg() > 0 {
		return Args{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if a.Commit != "" && !commitRe.MatchString(a.Commit) {
		return Args{}, fmt.Errorf("--commit must be 40 lowercase hex digits")
	}
	return a, nil
}
