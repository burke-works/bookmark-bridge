package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseNetscape_FlatAndNestedFolders(t *testing.T) {
	const input = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://example.com" ADD_DATE="1609459200">Example</A>
    <DT><H3>Reading List</H3>
    <DL><p>
        <DT><A HREF="https://a.example.com">A</A>
        <DT><A HREF="https://b.example.com" ADD_DATE="1612137600">B</A>
    </DL><p>
</DL><p>
`

	got, err := ParseNetscape(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseNetscape: %v", err)
	}

	want := &Folder{
		Title: "Bookmarks",
		Bookmarks: []Bookmark{
			{Title: "Example", URL: "https://example.com", AddedAt: "2021-01-01T00:00:00Z"},
		},
		Folders: []*Folder{
			{
				Title: "Reading List",
				Bookmarks: []Bookmark{
					{Title: "A", URL: "https://a.example.com"},
					{Title: "B", URL: "https://b.example.com", AddedAt: "2021-02-01T00:00:00Z"},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseNetscape mismatch\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestParseNetscape_UnescapesEntities(t *testing.T) {
	const input = `<DL><p>
    <DT><A HREF="https://example.com/?a=1&amp;b=2">Tom &amp; Jerry</A>
</DL><p>
`
	got, err := ParseNetscape(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseNetscape: %v", err)
	}
	if len(got.Bookmarks) != 1 {
		t.Fatalf("got %d bookmarks, want 1", len(got.Bookmarks))
	}
	b := got.Bookmarks[0]
	if b.Title != "Tom & Jerry" {
		t.Errorf("Title = %q, want %q", b.Title, "Tom & Jerry")
	}
	if b.URL != "https://example.com/?a=1&b=2" {
		t.Errorf("URL = %q, want %q", b.URL, "https://example.com/?a=1&b=2")
	}
}

func TestParseNetscape_NoAddDate(t *testing.T) {
	const input = `<DL><p>
    <DT><A HREF="https://example.com">Example</A>
</DL><p>
`
	got, err := ParseNetscape(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseNetscape: %v", err)
	}
	if len(got.Bookmarks) != 1 {
		t.Fatalf("got %d bookmarks, want 1", len(got.Bookmarks))
	}
	if got.Bookmarks[0].AddedAt != "" {
		t.Errorf("AddedAt = %q, want empty", got.Bookmarks[0].AddedAt)
	}
}

func TestWriteNetscape_Format(t *testing.T) {
	root := &Folder{
		Title: "Bookmarks",
		Bookmarks: []Bookmark{
			{Title: "Example", URL: "https://example.com"},
		},
		Folders: []*Folder{
			{
				Title: "Sub",
				Bookmarks: []Bookmark{
					{Title: "Inner", URL: "https://inner.example.com"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteNetscape(&buf, root); err != nil {
		t.Fatalf("WriteNetscape: %v", err)
	}

	want := netscapeHeader +
		"<DL><p>\n" +
		"    <DT><A HREF=\"https://example.com\">Example</A>\n" +
		"    <DT><H3>Sub</H3>\n" +
		"    <DL><p>\n" +
		"        <DT><A HREF=\"https://inner.example.com\">Inner</A>\n" +
		"    </DL><p>\n" +
		"</DL><p>\n"

	if buf.String() != want {
		t.Fatalf("WriteNetscape output mismatch\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestNetscape_RoundTrip(t *testing.T) {
	original := &Folder{
		Title: "Bookmarks",
		Bookmarks: []Bookmark{
			{Title: `Tom & Jerry "Show"`, URL: "https://example.com/?a=1&b=2", AddedAt: "2021-01-01T00:00:00Z"},
		},
		Folders: []*Folder{
			{
				Title: "Reading List",
				Bookmarks: []Bookmark{
					{Title: "A", URL: "https://a.example.com"},
					{Title: "B", URL: "https://b.example.com", AddedAt: "2021-02-01T00:00:00Z"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteNetscape(&buf, original); err != nil {
		t.Fatalf("WriteNetscape: %v", err)
	}

	roundTripped, err := ParseNetscape(&buf)
	if err != nil {
		t.Fatalf("ParseNetscape: %v", err)
	}

	if !reflect.DeepEqual(original, roundTripped) {
		t.Fatalf("round trip mismatch\noriginal:     %+v\nroundTripped: %+v", original, roundTripped)
	}
}
