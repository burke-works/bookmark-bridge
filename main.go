package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	from := flag.String("from", "netscape", "input format (netscape or json)")
	to := flag.String("to", "json", "output format (json or netscape)")
	in := flag.String("in", "-", "input file, or - for stdin")
	out := flag.String("out", "-", "output file, or - for stdout")
	flag.Parse()

	if err := run(*from, *to, *in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "bookmark-bridge:", err)
		os.Exit(1)
	}
}

func run(from, to, in, out string) error {
	r, closeIn, err := openInput(in)
	if err != nil {
		return err
	}
	defer closeIn()

	w, closeOut, err := openOutput(out)
	if err != nil {
		return err
	}
	defer closeOut()

	var root *Folder
	switch from {
	case "netscape":
		root, err = ParseNetscape(r)
		if err != nil {
			return fmt.Errorf("parsing bookmarks: %w", err)
		}
	case "json":
		root = &Folder{}
		if err := json.NewDecoder(r).Decode(root); err != nil {
			return fmt.Errorf("parsing json: %w", err)
		}
	default:
		return fmt.Errorf("unsupported input format %q (want netscape or json)", from)
	}

	switch to {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(root); err != nil {
			return fmt.Errorf("writing json: %w", err)
		}
	case "netscape":
		if err := WriteNetscape(w, root); err != nil {
			return fmt.Errorf("writing bookmarks: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output format %q (want netscape or json)", to)
	}
	return nil
}

func openInput(path string) (io.Reader, func() error, error) {
	if path == "" || path == "-" {
		return os.Stdin, func() error { return nil }, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening input: %w", err)
	}
	return f, f.Close, nil
}

func openOutput(path string) (io.Writer, func() error, error) {
	if path == "" || path == "-" {
		return os.Stdout, func() error { return nil }, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("creating output: %w", err)
	}
	return f, f.Close, nil
}
