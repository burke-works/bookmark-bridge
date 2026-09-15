package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	from := flag.String("from", "netscape", "input format (netscape)")
	to := flag.String("to", "json", "output format (json)")
	in := flag.String("in", "-", "input file, or - for stdin")
	out := flag.String("out", "-", "output file, or - for stdout")
	flag.Parse()

	if err := run(*from, *to, *in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "bookmark-bridge:", err)
		os.Exit(1)
	}
}

func run(from, to, in, out string) error {
	if from != "netscape" || to != "json" {
		return fmt.Errorf("unsupported conversion %q -> %q (only netscape -> json is implemented so far)", from, to)
	}

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

	root, err := ParseNetscape(r)
	if err != nil {
		return fmt.Errorf("parsing bookmarks: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(root); err != nil {
		return fmt.Errorf("writing json: %w", err)
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
