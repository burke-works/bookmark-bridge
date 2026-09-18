package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Every major browser (Chrome, Firefox, Safari) exports and imports
// bookmarks in this same loose, pre-XHTML dialect of HTML, defined
// originally by Netscape. It is not well-formed markup: <DT> and <p>
// tags are never closed. In practice every exporter puts one tag per
// line, so a line-oriented parser is enough and is far simpler than
// trying to coax a real HTML parser into accepting it.
var (
	reLink    = regexp.MustCompile(`(?i)<DT><A\s+([^>]*)>(.*?)</A>`)
	reFolder  = regexp.MustCompile(`(?i)<DT><H3[^>]*>(.*?)</H3>`)
	reHref    = regexp.MustCompile(`(?i)HREF="([^"]*)"`)
	reAddDate = regexp.MustCompile(`(?i)ADD_DATE="([^"]*)"`)
)

// ParseNetscape reads a Netscape-format bookmark export and returns the
// bookmark tree. The returned root folder has no bookmarks of its own
// beyond what appeared before the first sub-folder in the file.
func ParseNetscape(r io.Reader) (*Folder, error) {
	root := &Folder{Title: "Bookmarks"}
	stack := []*Folder{root}
	var pending *Folder // folder whose <H3> we've seen but whose <DL> hasn't opened yet

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if m := reLink.FindStringSubmatch(line); m != nil {
			attrs, title := m[1], html.UnescapeString(m[2])
			b := Bookmark{Title: title}
			if hm := reHref.FindStringSubmatch(attrs); hm != nil {
				b.URL = html.UnescapeString(hm[1])
			}
			if dm := reAddDate.FindStringSubmatch(attrs); dm != nil {
				if ts, err := strconv.ParseInt(dm[1], 10, 64); err == nil {
					b.AddedAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
				}
			}
			cur := stack[len(stack)-1]
			cur.Bookmarks = append(cur.Bookmarks, b)
			continue
		}

		if m := reFolder.FindStringSubmatch(line); m != nil {
			f := &Folder{Title: html.UnescapeString(m[1])}
			cur := stack[len(stack)-1]
			cur.Folders = append(cur.Folders, f)
			pending = f
			continue
		}

		upper := strings.ToUpper(line)
		switch {
		case strings.Contains(upper, "<DL>"):
			// The very first <DL> opens the root's own list, which is
			// already on the stack, so only push when a folder is waiting.
			if pending != nil {
				stack = append(stack, pending)
				pending = nil
			}
		case strings.Contains(upper, "</DL>"):
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading bookmarks: %w", err)
	}
	return root, nil
}

const netscapeHeader = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
`

// WriteNetscape writes root back out in Netscape Bookmark File format,
// the same dialect ParseNetscape reads. The round trip isn't lossless:
// AddedAt survives (as ADD_DATE, re-encoded to a Unix timestamp), but
// there's nowhere in our JSON shape yet to carry favicons or folder
// timestamps, so those are simply absent on the way back out.
func WriteNetscape(w io.Writer, root *Folder) error {
	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString(netscapeHeader); err != nil {
		return err
	}
	if err := writeNetscapeFolder(bw, root, 0); err != nil {
		return err
	}
	return bw.Flush()
}

func writeNetscapeFolder(w *bufio.Writer, f *Folder, depth int) error {
	indent := strings.Repeat("    ", depth)
	if _, err := fmt.Fprintf(w, "%s<DL><p>\n", indent); err != nil {
		return err
	}
	childIndent := indent + "    "

	for _, b := range f.Bookmarks {
		attrs := fmt.Sprintf(` HREF="%s"`, html.EscapeString(b.URL))
		if b.AddedAt != "" {
			if t, err := time.Parse(time.RFC3339, b.AddedAt); err == nil {
				attrs += fmt.Sprintf(` ADD_DATE="%d"`, t.Unix())
			}
		}
		if _, err := fmt.Fprintf(w, "%s<DT><A%s>%s</A>\n", childIndent, attrs, html.EscapeString(b.Title)); err != nil {
			return err
		}
	}

	for _, sub := range f.Folders {
		if _, err := fmt.Fprintf(w, "%s<DT><H3>%s</H3>\n", childIndent, html.EscapeString(sub.Title)); err != nil {
			return err
		}
		if err := writeNetscapeFolder(w, sub, depth+1); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s</DL><p>\n", indent); err != nil {
		return err
	}
	return nil
}
