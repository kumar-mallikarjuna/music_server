package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strings"
	"time"
)

// MongoDB Client
var Client *mongo.Client

// GraphQL schema
var schema graphql.Schema

func main() {
	var err error
	var resO MusicDoc

	schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: QueryType,
		},
	)

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	Client, err = mongo.Connect(context.TODO(), clientOptions)

	Scan("/home/legendrian/Music/Breaking Benjamin/")

	/*
		db.music.update({}, { $pull: { "artists.$[].albums.$[].tracks": {tname: "Without You"} } })
		db.music.update({}, { $pull: { "artists.$[].albums": {_id: "Dear Agony (Japanese Edition)"} } })
	*/

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := Client.Database("home_server").Collection("music")
	err = col.FindOne(ctx, bson.D{}).Decode(&resO)

	if err != nil {
		fmt.Println(err)
	}

	// Set ArtistsJ from DB (first document)
	ArtistsJ = resO.ArtistsB

	if err != nil {
		fmt.Println(err)
	}

	// Static (Songs) File Server
	fs := http.FileServer(http.Dir("/home/legendrian/Music/"))

	// Default endpoint
	http.HandleFunc("/", greet)
	// Endpoint for statics
	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	// Endpoint for ls
	http.HandleFunc("/ls/", Ls)
	// GraphQL API endpoint
	http.HandleFunc("/api/", api)

	http.ListenAndServe(":8080", nil)
}

// Default endpoint handler
func greet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, "Hey!")
}

// Neuter for Static File Server
func neuter(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// API Handler
func api(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	w.Header().Set("Content-Type", "application/json")

	if len(result.Errors) > 0 {
		fmt.Fprint(w, `{"success": false}`)
		return
	}

	b, _ := json.Marshal(result)
	fmt.Fprint(w, string(b))
}
