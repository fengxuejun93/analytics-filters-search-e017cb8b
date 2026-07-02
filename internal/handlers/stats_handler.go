package handlers

import "net/http"

func StatsAPIHandler(w http.ResponseWriter, r *http.Request) {
	stats := statsSvc.GetStatistics()
	renderJSON(w, stats)
}
