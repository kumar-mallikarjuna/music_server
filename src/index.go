package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/graphql-go/graphql"
)

// Track struct
type Track struct {
	Tno   string `json:"tno" bson:"tno"`
	Tname string `json:"tname" bson:"tname"`
	Tdur  string `json:"tdur" bson:"tdur"`
	Path  string `json:"path" bson:"path"`
}

// Track type (GraphQL)
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

// Album struct
type Album struct {
	Name   string  `json:"name" bson:"_id"`
	Genre  string  `json:"genre" bson:"genre"`
	Tracks []Track `json:"tracks" bson:"tracks"`
}

// Album type (GraphQL)
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

// Artists map
type Artist_T struct {
	Albums map[string]*Album `json:"albums"`
}

// Artists JSON/BSON struct
type ArtistJ struct {
	Name   string  `json:"name" bson:"_id"`
	Albums []Album `json:"albums" bson:"albums"`
}

// Artists GraphQL type
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

// Global map of Artists
var Artists map[string]Artist_T

// Global array of Artists (for JSON/BSON)
var ArtistsJ []ArtistJ

// Scan path to define Artists/ArtistsJ
func Scan(path string) {
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
			tno, _ := fInfo[0].GetString("TrackNumber")

			tname, _ := fInfo[0].GetString("Title")
			tdur, _ := fInfo[0].GetString("Duration")
			artist, _ := fInfo[0].GetString("Artist")
			album, _ := fInfo[0].GetString("Album")

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

	UpdateIndex(path)
}

// Query Fields for GraphQL interface
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

// Scan specified path (returns JSON)
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
