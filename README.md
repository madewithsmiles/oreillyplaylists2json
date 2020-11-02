# oreillyplaylists2json
[Playing with Go] - Small weekend project to extract your Safari (O'Reilly Online Learning) playlists HTML pages into a single JSON file

## Limitations
*Works offline only.\
*Provided as is, and needs improvement like accepting cmd line input, better and online parsing of the html..etc \
**Goal was to keep things simple to learn/play with Go ;)**

## How it works
Goes through current directory and parses each HTML file and merge results in a json file. Useful to easily share/move your playlists with a single json. Notice: you still need O'Reilly to access the book/video/etc..  

## How to use:
- For each of your playlists on Safari [Sep 8 2020 - the format might have changed :(]:
  - open the developer console in Chrome, paste and run:\
    `copy(document.getElementsByClassName("viewMain--u11Is")[0].outerHTML)`
  - create an html file in your directory of choice and paste the content
- Run: `go run playlistextractor.go` to get a file named `all_playlists.json` containing all of your playlists items in your cwd :)

