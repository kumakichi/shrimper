package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defType     string = ""
	baseDir     string = "."
	version     string = "0.2.1"
	description string = "Xiami music downloader."
	prog        string = "shrimper"
)

var (
	dlType      string
	id          int
	dir         string
	showVersion bool
	noTag       bool
)

func init() {
	flag.StringVar(&dlType, "t", defType, "single / album / omnibus")
	flag.IntVar(&id, "i", -1, "id in the url")
	flag.StringVar(&dir, "d", baseDir, "specify the directory to store files;")
	flag.BoolVar(&showVersion, "v", false, "show program's version number and exit")
	flag.BoolVar(&noTag, "nt", false, "skip adding ID3 tag")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-h/--help] [-nt] [-d DIRECTORY] [-v] -t type -i id\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s\n\n", description)
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	if showVersion {
		fmt.Printf("%s %s\n", prog, version)
		return
	}

	//var tracklist []track = make([]track, 0)
	url := gen_tracklist_urls(dlType, id)
	tracklist := get_tracklist_from_url(url)

	for i := 0; i < len(tracklist); i++ {
		tracklist[i].Location = decrypt_location(tracklist[i].Location)
	}
	fmt.Printf("%d file(s) to download\n", len(tracklist))

	for i, t := range tracklist {
		dest_file_path := fmt.Sprintf("%s/%s-%s.mp3", baseDir, t.Title, t.Artist)
		download(t.Location, dest_file_path)
		if !noTag {
			add_id3_tag(dest_file_path, t)
		}
		progress := fmt.Sprintf("[%d/%d] COMPLETE | %s\n", i+1, len(tracklist), dest_file_path)
		fmt.Println(progress)
	}
}

func download(url, filename string) {
	absPath, err := filepath.Abs(filename)
	parent_dir := filepath.Dir(absPath)
	errExit(err)

	_, err = os.Stat(parent_dir)
	if os.IsNotExist(err) {
		os.Mkdir(parent_dir, os.ModePerm)
	}

	resp, err := http.Get(url)
	errExit(err)
	defer resp.Body.Close()

	file, err := os.Create(absPath)
	errExit(err)
	defer file.Close()

	io.Copy(file, resp.Body)
}
