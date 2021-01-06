package main

import "net/http"
import "fmt"
import "strings"

var Root FSNode

func main() {
	fs := http.FileServer(http.Dir("/home/legendrian/Music/"))

	http.HandleFunc("/", greet)
	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	http.HandleFunc("/ls/", Ls)

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
