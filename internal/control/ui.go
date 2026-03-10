package control

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/index.html web/assets/app.css web/assets/app.js web/assets/app-install.js web/assets/app-render.js web/assets/app-maintenance.js web/assets/app-product.js
var uiFS embed.FS

type uiAsset struct {
	path        string
	contentType string
}

var uiAssets = map[string]uiAsset{
	"/":                          {path: "web/index.html", contentType: "text/html; charset=utf-8"},
	"/index.html":                {path: "web/index.html", contentType: "text/html; charset=utf-8"},
	"/assets/app.css":            {path: "web/assets/app.css", contentType: "text/css; charset=utf-8"},
	"/assets/app.js":             {path: "web/assets/app.js", contentType: "application/javascript; charset=utf-8"},
	"/assets/app-install.js":     {path: "web/assets/app-install.js", contentType: "application/javascript; charset=utf-8"},
	"/assets/app-render.js":      {path: "web/assets/app-render.js", contentType: "application/javascript; charset=utf-8"},
	"/assets/app-maintenance.js": {path: "web/assets/app-maintenance.js", contentType: "application/javascript; charset=utf-8"},
	"/assets/app-product.js":     {path: "web/assets/app-product.js", contentType: "application/javascript; charset=utf-8"},
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset, ok := uiAssets[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	body, err := fs.ReadFile(uiFS, asset.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", asset.contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
