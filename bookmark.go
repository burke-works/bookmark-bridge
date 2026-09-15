package main

// Bookmark is a single saved link, plus whatever metadata survived the
// round trip from whatever format it came from.
type Bookmark struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	AddedAt string `json:"added_at,omitempty"` // RFC3339, when the source format has it
}

// Folder is a node in the bookmark tree. The root folder represents the
// whole bookmark file; its Title is usually just "Bookmarks".
type Folder struct {
	Title     string    `json:"title"`
	Bookmarks []Bookmark `json:"bookmarks,omitempty"`
	Folders   []*Folder  `json:"folders,omitempty"`
}
