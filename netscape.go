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
