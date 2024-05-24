package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

type responseBody struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

func run(ctx context.Context) error {
	router := http.NewServeMux()

	router.HandleFunc("GET /latest", func(w http.ResponseWriter, r *http.Request) {
		body := responseBody{
			Version: "v0.0.2",
			Path:    "v0.0.2-app.tar.gz",
		}
		json.NewEncoder(w).Encode(body)
	})

	router.HandleFunc("GET /file", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		http.ServeFile(w, r, path)
	})

	fmt.Println("listing on :3000")
	return http.ListenAndServe(":3000", router)
}
