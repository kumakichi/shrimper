package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Playlist struct {
	XMLName   xml.Name  `xml:"playlist"`
	Tracklist tracklist `xml:"trackList"`
}

type tracklist struct {
	XMLName xml.Name `xml:"trackList"`
	Tracks  []track  `xml:"track"`
}

type track struct {
	XMLName    xml.Name `xml:"track"`
	Title      string   `xml:"title"`
	Album_name string   `xml:"album_name"`
	Artist     string   `xml:"artist"`
	Location   string   `xml:"location"`
	Pic        string   `xml:"pic"`
}

// Calc the url depending the ids of songs, albums or omnibuses.
func gen_tracklist_urls(url_type string, id int) string {
	var template string
	url_base := "http://www.xiami.com/song/playlist/id/"
	url_single := "/object_name/default/object_id/0"
	url_album := "/type/1"
	url_omnibus := "/type/3"

	switch url_type {
	case "single":
		template = url_single
	case "album":
		template = url_album
	case "omnibus":
		template = url_omnibus
	default:
		fmt.Println("invalid url_type")
		os.Exit(-1)
	}

	return fmt.Sprintf("%s%d%s", url_base, id, template)
}

// Get tracklist from url
func get_tracklist_from_url(url string) []track {
	raw := get_response(url)
	if raw == nil {
		fmt.Printf("ERRROR: Nothing in this URL: %s\n", url)
		os.Exit(1)
	}

	tracklists := make([]track, 0)
	p := Playlist{}
	err := xml.Unmarshal(raw, &p)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil
	}

	for _, t := range p.Tracklist.Tracks {
		tracklists = append(tracklists, t)
	}
	return tracklists
}

// Abtain the xml structure in byte string.
// If you push a request without the headers, there may trigger 503/403 error.
func get_response(url string) []byte {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 7.1; Trident/5.0)",
		"Referer":    "http://www.xiami.com/song/play",
	}

	request, err := http.NewRequest("GET", url, nil)
	errExit(err)

	for key, val := range headers {
		request.Header.Set(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	errExit(err)
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	errExit(err)

	return contents
}

// Analyze the track's location
func decrypt_location(location string) string {
	urlStr := location[1:]
	urllen := len(urlStr)
	row_number := int(location[0]) - '0'

	col_number := urllen / row_number // basic column count: 34
	spare_char := urllen % row_number // count of rows that have 1 more column

	var length int
	matrix := make([]string, 0)
	for i := 0; i < row_number; i++ {
		if i < spare_char {
			length = col_number + 1
		} else {
			length = col_number
		}

		matrix = append(matrix, urlStr[:length])
		urlStr = urlStr[length:]
	}

	urlStr = ""
	for i := 0; i < urllen; i++ {
		urlStr += string(matrix[i%row_number][i/row_number])
	}

	s, _ := url.QueryUnescape(urlStr)
	return strings.Replace(s, "^", "0", -1)
}

// Get album image url in a specific size:
//     * Leave None for the largest
//     * 4 for a reasonable size
func gen_album_image_url(basic string, size int) string {
	if size < 1 || size > 4 {
		return basic
	}
	img_url := fmt.Sprintf("%s%d%s", basic[:len(basic)-5], size, basic[len(basic)-4:])
	return img_url
}

// Write id3v2.4 tag to mp3 file.
func add_id3_tag(filename string, t track) {
}

func errExit(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
