package main

import (
	"encoding/json"
	"fmt"
	"github.com/barasher/go-exiftool"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Track struct {
	Tno   string `json:"tno"`
	Tname string `json:"tname"`
	Tdur  string `json:"tdur"`
	Path  string `json:"path"`
}

type Album struct {
	Genre  string  `json:"genre"`
	Tracks []Track `json:"tracks"`
}

type Artist struct {
	Albums map[string]*Album `json:"albums"`
}

var Artists map[string]Artist

func scan(path string) {
	// path := "/home/legendrian/Music/Arctic Monkeys"
	Artists = make(map[string]Artist)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		et, err := exiftool.NewExiftool()

		if err != nil {
			fmt.Println(err)
			return err
		}

		defer et.Close()

		fInfo := et.ExtractMetadata(path)
		mime_t, ok := fInfo[0].Fields["MIMEType"].(string)

		if ok && strings.HasPrefix(mime_t, "audio") {
			tno, _ := fInfo[0].Fields["TrackNumber"].(string)

			tname, _ := fInfo[0].Fields["Title"].(string)
			tdur, _ := fInfo[0].Fields["Duration"].(string)
			artist, _ := fInfo[0].Fields["Artist"].(string)
			album, _ := fInfo[0].Fields["Album"].(string)

			if _, ok := Artists[artist]; !ok {
				Artists[artist] = Artist{map[string]*Album{}}
			}

			if _, ok := Artists[artist].Albums[album]; !ok {
				genre := fInfo[0].Fields["Genre"].(string)
				Artists[artist].Albums[album] = &Album{genre, []Track{}}
				Artists[artist].Albums[album].Genre = genre
			}

			Artists[artist].Albums[album].Tracks = append(Artists[artist].Albums[album].Tracks, Track{tno, tname, tdur, path})
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	b, err := json.Marshal(Artists["Arctic Monkeys"])

	if err == nil {
		fmt.Printf("%s\n", b)
	}
}

func Ls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		loc, err := url.PathUnescape(r.URL.EscapedPath())
		ls, err := ioutil.ReadDir("/home/legendrian/Music/" + loc[4:])
		fnames := []string{}

		if err != nil {
			log.Panic(err)
		}

		for _, x := range ls {
			fnames = append(fnames, x.Name())
		}

		fmt.Println(fnames)

		out := map[string][]string{"names": fnames}
		outB, _ := json.Marshal(out)
		w.Write(outB)
	}
}
