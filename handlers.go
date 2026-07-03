package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /api/items
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	filter := FilterParams{
		Keyword:  r.URL.Query().Get("keyword"),
		Category: r.URL.Query().Get("category"),
		City:     r.URL.Query().Get("city"),
		Status:   r.URL.Query().Get("status"),
	}
	result := h.store.ListItemsWithTotal(filter)
	writeJSON(w, http.StatusOK, result)
}

// GET /api/items/{id}
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request, id string) {
	item, ok := h.store.GetItem(id)
	if !ok {
		writeError(w, http.StatusNotFound, "货品不存在")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// GET /api/items/{id}/detail (聚合详情)
func (h *Handler) GetItemDetail(w http.ResponseWriter, r *http.Request, id string) {
	detail, ok := h.store.GetItemDetail(id)
	if !ok {
		writeError(w, http.StatusNotFound, "货品不存在")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// GET /api/items/{id}/history
func (h *Handler) GetStatusHistory(w http.ResponseWriter, r *http.Request, itemID string) {
	history := h.store.GetStatusHistory(itemID)
	writeJSON(w, http.StatusOK, history)
}

// POST /api/items
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Title == "" || req.Category == "" || req.Publisher == "" {
		writeError(w, http.StatusBadRequest, "标题、品类、发布人为必填项")
		return
	}
	item := h.store.CreateItem(req)
	writeJSON(w, http.StatusCreated, item)
}

// PUT /api/items/{id}
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request, id string) {
	var req ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Title == "" || req.Category == "" || req.Publisher == "" {
		writeError(w, http.StatusBadRequest, "标题、品类、发布人为必填项")
		return
	}
	item, err := h.store.UpdateItem(id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// PUT /api/items/{id}/status
func (h *Handler) UpdateItemStatus(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	item, err := h.store.UpdateItemStatus(id, ItemStatus(body.Status))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// GET /api/items/{id}/applications
func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request, itemID string) {
	apps := h.store.ListApplications(itemID)
	writeJSON(w, http.StatusOK, apps)
}

// POST /api/items/{id}/applications
func (h *Handler) CreateApplication(w http.ResponseWriter, r *http.Request, itemID string) {
	var req ApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Applicant == "" {
		writeError(w, http.StatusBadRequest, "申请人为必填项")
		return
	}
	app, err := h.store.CreateApplication(itemID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

// PUT /api/applications/{id}
func (h *Handler) HandleApplication(w http.ResponseWriter, r *http.Request, appID string) {
	var req ApplicationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	app, err := h.store.HandleApplication(appID, req.Action)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, app)
}

// GET /api/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetStats()
	writeJSON(w, http.StatusOK, stats)
}

// GET /api/filters
func (h *Handler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": h.store.GetCategories(),
		"cities":     h.store.GetCities(),
		"statuses": []map[string]string{
			{"value": "listed", "label": "上架中"},
			{"value": "exchanged", "label": "已置换"},
			{"value": "delisted", "label": "已下架"},
		},
	})
}

// 注册路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListItems(w, r)
		case http.MethodPost:
			h.CreateItem(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/api/items/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/items/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 && parts[0] != "" {
			itemID := parts[0]
			switch r.Method {
			case http.MethodGet:
				h.GetItem(w, r, itemID)
			case http.MethodPut:
				h.UpdateItem(w, r, itemID)
			default:
				writeError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
			return
		}

		if len(parts) == 2 && parts[0] != "" {
			itemID := parts[0]
			sub := parts[1]
			switch sub {
			case "status":
				if r.Method == http.MethodPut {
					h.UpdateItemStatus(w, r, itemID)
				} else {
					writeError(w, http.StatusMethodNotAllowed, "方法不允许")
				}
			case "detail":
				if r.Method == http.MethodGet {
					h.GetItemDetail(w, r, itemID)
				} else {
					writeError(w, http.StatusMethodNotAllowed, "方法不允许")
				}
			case "history":
				if r.Method == http.MethodGet {
					h.GetStatusHistory(w, r, itemID)
				} else {
					writeError(w, http.StatusMethodNotAllowed, "方法不允许")
				}
			case "applications":
				switch r.Method {
				case http.MethodGet:
					h.ListApplications(w, r, itemID)
				case http.MethodPost:
					h.CreateApplication(w, r, itemID)
				default:
					writeError(w, http.StatusMethodNotAllowed, "方法不允许")
				}
			default:
				writeError(w, http.StatusNotFound, "路径不存在")
			}
			return
		}

		writeError(w, http.StatusNotFound, "路径不存在")
	})

	mux.HandleFunc("/api/applications/", func(w http.ResponseWriter, r *http.Request) {
		appID := strings.TrimPrefix(r.URL.Path, "/api/applications/")
		if appID == "" {
			writeError(w, http.StatusNotFound, "申请ID不能为空")
			return
		}
		if r.Method == http.MethodPut {
			h.HandleApplication(w, r, appID)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetStats(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/api/filters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetFilterOptions(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// SPA 回退
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		http.NotFound(w, r)
	})
}

// 辅助：从查询参数获取 int
func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
