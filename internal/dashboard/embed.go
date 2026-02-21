package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var assets embed.FS

func SPAHandler() http.Handler {
	fsys, _ := fs.Sub(assets, "dist")
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		_, err := fsys.Open(path)
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
