package main

import (
	"encoding/json"
	"fmt"
	"github.com/barasher/go-exiftool"
	"github.com/graphql-go/graphql"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Track struct {
	Tno   string `json:"tno" bson:"tno"`
	Tname string `json:"tname" bson:"tname"`
	Tdur  string `json:"tdur" bson:"tdur"`
	Path  string `json:"path" bson:"path"`
}

var TrackG = graphql.NewObject(graphql.ObjectConfig{
	Name: "Track",
	Fields: graphql.Fields{
		"tno": &graphql.Field{
			Type: graphql.String,
		},
		"tname": &graphql.Field{
			Type: graphql.String,
		},
		"tdu": &graphql.Field{
			Type: graphql.String,
		},
		"path": &graphql.Field{
			Type: graphql.String,
		},
	},
})

type Album struct {
	Name   string  `json:"name" bson:"name"`
	Genre  string  `json:"genre" bson:"genre"`
	Tracks []Track `json:"tracks" bson:"tracks"`
}

var AlbumG = graphql.NewObject(graphql.ObjectConfig{
	Name: "Album",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"genre": &graphql.Field{
			Type: graphql.String,
		},
		"tracks": &graphql.Field{
			Type: graphql.NewList(TrackG),
		},
	},
})

type Artist_T struct {
	Albums map[string]*Album `json:"albums"`
}

type ArtistJ struct {
	Name   string  `json:"name" bson:"name"`
	Albums []Album `json:"albums" bson:"albums"`
}

var ArtistG = graphql.NewObject(graphql.ObjectConfig{
	Name: "Artist",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"albums": &graphql.Field{
			Type: graphql.NewList(AlbumG),
		},
	},
})

var Artists map[string]Artist_T
var ArtistsJ []ArtistJ

func scan(path string) {
	// path := "/home/legendrian/Music/Arctic Monkeys"
	Artists = make(map[string]Artist_T)

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

		if info.IsDir() {
			fmt.Println(path)
		}

		fInfo := et.ExtractMetadata(path)
		mime_t, ok := fInfo[0].Fields["MIMEType"].(string)

		if ok && strings.HasPrefix(mime_t, "audio") {
			tno, _ := fInfo[0].Fields["TrackNumber"].(string)

			tname, _ := fInfo[0].Fields["Title"].(string)
			tdur, _ := fInfo[0].Fields["Duration"].(string)
			artist, _ := fInfo[0].Fields["Artist"].(string)
			album, _ := fInfo[0].Fields["Album"].(string)

			if _, ok := Artists[artist]; !ok {
				Artists[artist] = Artist_T{map[string]*Album{}}
			}

			if _, ok := Artists[artist].Albums[album]; !ok {
				genre := fInfo[0].Fields["Genre"].(string)
				Artists[artist].Albums[album] = &Album{album, genre, []Track{}}
				Artists[artist].Albums[album].Genre = genre
			}

			Artists[artist].Albums[album].Tracks = append(Artists[artist].Albums[album].Tracks, Track{tno, tname, tdur, path})
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	for k, v := range Artists {
		var albumsj []Album

		for _, va := range v.Albums {
			albumsj = append(albumsj, *va)
		}

		ArtistsJ = append(ArtistsJ, ArtistJ{k, albumsj})
	}
}

var QueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"list": &graphql.Field{
			Type: graphql.NewList(ArtistG),
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return ArtistsJ, nil
			},
		},
	},
})

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
