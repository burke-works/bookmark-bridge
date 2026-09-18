# bookmark-bridge

Browser bookmark exports are all stuck in the same format: the Netscape
Bookmark File, a dialect of HTML from the mid-90s with unclosed `<DT>` and
`<p>` tags, folders represented as nested `<H3>`/`<DL>` pairs, and dates
stored as raw Unix timestamps in an `ADD_DATE` attribute. It's what Chrome,
Firefox, and Safari all still produce when you hit "Export Bookmarks", and
it's miserable to script against.

`bookmark-bridge` reads that format and turns it into plain, nested JSON so
you can grep it, diff it, or feed it into something else. It also goes the
other way, turning that JSON back into a Netscape file a browser can import.

## Usage

Build it:

```
go build -o bookmark-bridge .
```

Convert a file:

```
./bookmark-bridge -in bookmarks.html -out bookmarks.json
```

Or pipe it through stdin/stdout, which works the same way:

```
cat bookmarks.html | ./bookmark-bridge > bookmarks.json
```

Both `-in` and `-out` default to `-`, meaning stdin and stdout, so the two
commands above behave identically. Mixing the two is fine too:

```
./bookmark-bridge -in bookmarks.html > bookmarks.json
```

Convert back the other way with `-from json -to netscape`:

```
./bookmark-bridge -from json -to netscape -in bookmarks.json -out bookmarks.html
```

## Output shape

Each folder becomes a JSON object with its own bookmarks and sub-folders:

```json
{
  "title": "Bookmarks",
  "folders": [
    {
      "title": "Reading List",
      "bookmarks": [
        {
          "title": "Example Domain",
          "url": "https://example.com",
          "added_at": "2024-03-14T09:00:00Z"
        }
      ]
    }
  ]
}
```

`added_at` is only present when the source file had an `ADD_DATE`.

## Current limitations

The JSON -> Netscape direction only round-trips what the JSON shape above can
hold, so favicons and folder-level ADD_DATE/LAST_MODIFIED metadata (which
Netscape files can carry but this tool doesn't parse into JSON) are dropped.
There's also no support yet for Chrome/Firefox's internal JSON bookmark
formats, which differ from the JSON shape produced here.

## License

MIT, see LICENSE.
