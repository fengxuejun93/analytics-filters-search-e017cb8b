package router

import (
	"net/http"
	"path/filepath"

	"secondhand-exchange/internal/handlers"
)

func NewRouter(templateDir string, staticDir string) (http.Handler, error) {
	if err := handlers.InitHandlers(templateDir); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("GET /", handlers.ItemListHandler)
	mux.HandleFunc("GET /items/{id}", handlers.ItemDetailHandler)
	mux.HandleFunc("GET /items/new", handlers.ItemNewFormHandler)
	mux.HandleFunc("POST /items/new", handlers.ItemCreateHandler)
	mux.HandleFunc("GET /items/{id}/edit", handlers.ItemEditFormHandler)
	mux.HandleFunc("POST /items/{id}/edit", handlers.ItemUpdateHandler)
	mux.HandleFunc("POST /items/{id}/delete", handlers.ItemDeleteHandler)

	mux.HandleFunc("POST /exchange/apply", handlers.ExchangeApplyHandler)
	mux.HandleFunc("POST /exchange/{id}/accept", handlers.ExchangeAcceptHandler)
	mux.HandleFunc("POST /exchange/{id}/reject", handlers.ExchangeRejectHandler)
	mux.HandleFunc("POST /exchange/{id}/cancel", handlers.ExchangeCancelHandler)

	mux.HandleFunc("GET /api/stats", handlers.StatsAPIHandler)

	return mux, nil
}

func GetTemplateDir(baseDir string) string {
	return filepath.Join(baseDir, "templates")
}

func GetStaticDir(baseDir string) string {
	return filepath.Join(baseDir, "static")
}

func GetDataDir(baseDir string) string {
	return filepath.Join(baseDir, "data")
}
