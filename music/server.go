package main

import "net/http"
import "fmt"
import "strings"

func main() {
	greet := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprint(w, "Hey!")
	}

	fs := http.FileServer(http.Dir("/home/legendrian/Music/"))

	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	http.HandleFunc("/", greet)

	http.ListenAndServe(":8080", nil)
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
