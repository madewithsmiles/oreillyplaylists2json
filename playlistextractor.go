package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type PlaylistItem struct {
	ItemName   string `json:"itemName"`
	ItemAuthor string `json:"itemAuthor"`
	ItemID     string `json:"itemId"`
	ItemType   string `json:"itemType"`
	ItemHref   string `json:"itemHref"`
}

type Playlist struct {
	Name          string         `json:"name"`
	Author        string         `json:"author"`
	Description   string         `json:"description"`
	LastModified  string         `json:"lastModified"`
	NumberOfItems string         `json:"numberOfItems"`
	Items         []PlaylistItem `json:"items"`
}

type PlatlistHTML struct {
	Data            string
	Key             string
	ResearchedValue string
	RegexForSearch  *regexp.Regexp
	Get             func(*html.Node) string
	GetSubmatch     func(*regexp.Regexp, string) string
}

var researchDict = map[string][]PlatlistHTML{
	"h1": {{Key: "class", ResearchedValue: "Name", RegexForSearch: regexp.MustCompile(`.*playlistTitle.*`), Get: func(n *html.Node) string { return n.FirstChild.Data }}},
	"p": {{Key: "class", ResearchedValue: "Author", RegexForSearch: regexp.MustCompile(`.*orm-ByLine-root byline-.*`), Get: func(n *html.Node) string {
		// re := regexp.MustCompile(`By\s(.*)`)
		// fmt.Println("Author", n.Data, n.Attr, re.FindStringSubmatch(GetAllSiblingsValues(n)))
		// return re.FindStringSubmatch()[1]
		return GetAllSiblingsValues(n)
	}}, {Key: "class", ResearchedValue: "Description", RegexForSearch: regexp.MustCompile(`.*description-.*`), Get: func(n *html.Node) string { return n.FirstChild.Data }}, {Key: "class", ResearchedValue: "ItemAuthor", RegexForSearch: regexp.MustCompile(`.*orm-ByLine-root orm-Card-byline.*`), Get: func(n *html.Node) string {
		// re := regexp.MustCompile(`By\s(.*)`)
		// fmt.Println("Description", n.Data, n.Attr, re.FindStringSubmatch(GetAllSiblingsValues(n)))
		// return re.FindStringSubmatch(GetAllSiblingsValues(n))[1]
		return GetAllSiblingsValues(n)
	}}},
	"span": {{Key: "class", ResearchedValue: "LastModified", RegexForSearch: regexp.MustCompile(`.*lastUpdatedDetails-.*`), Get: func(n *html.Node) string {
		// re := regexp.MustCompile(`Last Updated: (.*)`)
		// fmt.Println("LastModified", n.Data, n.Attr, re.FindStringSubmatch(GetAllSiblingsValues(n)))
		// return re.FindStringSubmatch()[1]
		return n.FirstChild.FirstChild.Data
	}}},
	"div": {{Key: "class", ResearchedValue: "NumberOfItems", RegexForSearch: regexp.MustCompile(`.*itemCount-.*`), Get: func(n *html.Node) string {
		// re := regexp.MustCompile(`(\d+) Items`)
		// fmt.Println("NumberOfItems", n.Data, n.Attr, re.FindStringSubmatch(GetAllSiblingsValues(n)))
		// return re.FindStringSubmatch(n.FirstChild.Data)[1]
		return n.FirstChild.Data
	}}},
	"a": {{Key: "href", ResearchedValue: "Item", RegexForSearch: regexp.MustCompile(`.*/(\d+).*`), Get: func(n *html.Node) string { return n.FirstChild.Data }, GetSubmatch: func(re *regexp.Regexp, featureHref string) string {
		// fmt.Println("Item", featureHref, re, re.FindStringSubmatch(featureHref))
		return re.FindStringSubmatch(featureHref)[1]
	}}},
}

var values []Playlist

func GetType(href string) string {
	reVideo := regexp.MustCompile(`.*/(videos|dashboard|library|learning-paths)/.*`)
	if matches := reVideo.FindStringSubmatch(href); len(matches) > 0 {
		switch matches[1] {
		case "dashboard":
			return "book"
		case "videos":
			return "video"
		case "library":
			return "book"
		case "learning-paths":
			return "learning-path"
		}
	}
	fmt.Println(href)
	panic("could not get the type")
}

func Clean(val string) string {
	return fmt.Sprintf("%v", strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(val), "<span>", ""), "</span>", "")), " "))
}

func GetAllSiblingsValues(node *html.Node) string {
	values := &[]string{}
	child := &html.Node{}
	child = node.FirstChild
	*values = append(*values, child.Data)
	for {

		child = child.NextSibling
		if child.FirstChild != nil {
			*values = append(*values, child.FirstChild.Data)
		}

		if child.NextSibling == nil {
			break
		}

	}

	return strings.Join(*values, " ")
}

func f(n *html.Node, playlist *Playlist) {

	if n.Type == html.ElementNode {
		if rightPlaylistHTMLArray, ok := researchDict[n.Data]; ok {
			for _, a := range n.Attr {
				for _, rightPlaylistHTML := range rightPlaylistHTMLArray {
					if a.Key == rightPlaylistHTML.Key {
						if rightPlaylistHTML.RegexForSearch.MatchString(a.Val) {
							value := rightPlaylistHTML.Get(n)
							switch rightPlaylistHTML.ResearchedValue {
							case "Name":
								playlist.Name = value
							case "Author":
								playlist.Author = Clean(value)
							case "Description":
								playlist.Description = value
							case "LastModified":
								playlist.LastModified = value
							case "NumberOfItems":
								playlist.NumberOfItems = value
							case "ItemAuthor":
								// fmt.Println("Here", n.Data)
								playlist.Items[len(playlist.Items)-1].ItemAuthor = Clean(value)
							case "Item":
								// fmt.Println(a.Key, "=", a.Val)
								aPlaylistItem := new(PlaylistItem)
								aPlaylistItem.ItemID = rightPlaylistHTML.GetSubmatch(rightPlaylistHTML.RegexForSearch, a.Val)
								aPlaylistItem.ItemName = Clean(value)
								aPlaylistItem.ItemType = GetType(a.Val)
								aPlaylistItem.ItemHref = "https://learning.oreilly.com" + a.Val
								playlist.Items = append(playlist.Items, *aPlaylistItem)
							}
						}
					}
				}

			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f(c, playlist)
	}
}

func main() {

	var (
		playlists []string
		isHTML    bool
		err       error
	)

	const htmlExt = ".html"
	root, err := os.Getwd()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		isHTML = strings.ToLower(filepath.Ext(info.Name())) == htmlExt

		if isHTML {
			playlists = append(playlists, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
		return
	}

	fmt.Println("Found ", len(playlists), "playlists!\n")

	var playlist *Playlist

	for _, cpath := range playlists {
		fmt.Println("Reading path " + cpath)

		doc, err := ioutil.ReadFile(cpath)
		fdoc, err := html.Parse(strings.NewReader(string(doc)))
		if err != nil {
			log.Fatal(err)
		}

		playlist = new(Playlist)
		f(fdoc, playlist)
		values = append(values, *playlist)

	}

	asjson, err := json.MarshalIndent(values, "", " ")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// fmt.Println(string(asjson))
	outputFileName := "./all_playlists.json"
	f, err := os.Create(outputFileName)
	if err != nil {
		os.Exit(2)
	}
	defer f.Close()

	f.WriteString(string(asjson))
	fmt.Println("The file", outputFileName[2:], "was created")
}
