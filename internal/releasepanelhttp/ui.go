package releasepanelhttp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web web/assets
var webFS embed.FS

func handleUI(w http.ResponseWriter, r *http.Request) {
	subtree, err := fs.Sub(webFS, "web")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Path == "/" {
		http.ServeFileFS(w, r, subtree, "index.html")
		return
	}
	http.FileServerFS(subtree).ServeHTTP(w, r)
}
