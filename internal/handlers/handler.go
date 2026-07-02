package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"secondhand-exchange/internal/models"
	"secondhand-exchange/internal/services"
)

var (
	tpl        *template.Template
	itemSvc    *services.ItemService
	exchangeSvc *services.ExchangeService
	statsSvc   *services.StatsService
)

type PageData struct {
	Title      string
	ActiveMenu string
	Stats      *models.Statistics
	Data       interface{}
	Error      string
	Success    string
}

func InitHandlers(templateDir string) error {
	var err error
	tpl, err = template.New("").Funcs(template.FuncMap{
		"formatTime": func(t interface{}) string {
			type formatter interface {
				Format(string) string
			}
			if ft, ok := t.(formatter); ok {
				return ft.Format("2006-01-02 15:04")
			}
			return ""
		},
		"statusText": func(s string) string {
			if text, ok := models.ItemStatusMap[s]; ok {
				return text
			}
			if text, ok := models.ExchangeStatusMap[s]; ok {
				return text
			}
			return s
		},
		"itemStatusClass": func(s string) string {
			switch s {
			case models.ItemStatusOnSale:
				return "status-success"
			case models.ItemStatusPending:
				return "status-warning"
			case models.ItemStatusCompleted:
				return "status-info"
			case models.ItemStatusOffline:
				return "status-dark"
			default:
				return "status-default"
			}
		},
		"exchangeStatusClass": func(s string) string {
			switch s {
			case models.ExchangeStatusPending:
				return "status-warning"
			case models.ExchangeStatusAccepted:
				return "status-success"
			case models.ExchangeStatusRejected:
				return "status-danger"
			case models.ExchangeStatusCanceled:
				return "status-dark"
			default:
				return "status-default"
			}
		},
		"contains": strings.Contains,
		"emojiForCategory": func(category string) string {
			emojiMap := map[string]string{
				"数码电子": "💻",
				"图书文具": "📚",
				"家居用品": "🛋️",
				"服饰鞋帽": "👟",
				"运动户外": "🚴",
				"母婴用品": "🍼",
				"其他":     "🎁",
			}
			if emoji, ok := emojiMap[category]; ok {
				return emoji
			}
			return "📦"
		},
	}).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return err
	}

	itemSvc = services.NewItemService()
	exchangeSvc = services.NewExchangeService()
	statsSvc = services.NewStatsService()

	return nil
}

func render(w http.ResponseWriter, name string, data *PageData) {
	if data.Stats == nil {
		data.Stats = statsSvc.GetStatistics()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getFormValue(r *http.Request, key string) string {
	return strings.TrimSpace(r.FormValue(key))
}

func getPathValue(r *http.Request, key string) string {
	return r.PathValue(key)
}
