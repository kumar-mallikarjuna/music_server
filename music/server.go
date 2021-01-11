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
	"os"
	"strings"
	"time"
)

var Client *mongo.Client
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

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := Client.Database("home_server").Collection("music")
	err = col.FindOne(ctx, bson.D{}).Decode(&resO)

	if err != nil {
		fmt.Println(err)
	}

	ArtistsJ = resO.ArtistsB

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fs := http.FileServer(http.Dir("/home/legendrian/Music/"))

	http.HandleFunc("/", greet)
	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	http.HandleFunc("/ls/", Ls)
	http.HandleFunc("/api/", api)

	http.ListenAndServe(":8080", nil)
}

func greet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, "Hey!")
}

func neuter(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

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
