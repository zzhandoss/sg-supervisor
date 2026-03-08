package releasepanelhttp

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/recipe", s.handleRecipeUpdate)
	mux.HandleFunc("/api/v1/upstream/versions", s.handleVersions)
	mux.HandleFunc("/api/v1/releases/local", s.handleLocalRelease)
	mux.HandleFunc("/api/v1/licenses/issue", s.handleIssueLicense)
	mux.HandleFunc("/", handleUI)
	return mux
}
