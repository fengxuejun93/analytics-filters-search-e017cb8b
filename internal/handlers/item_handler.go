package handlers

import (
	"net/http"

	"secondhand-exchange/internal/models"
)

type ListPageData struct {
	Items       []*models.Item
	Filter      *models.ItemFilter
	Categories  []string
	Cities      []string
	Statuses    map[string]string
}

type DetailPageData struct {
	Item        *models.Item
	Requests    []*models.ExchangeRequest
	CurrentUser string
}

type FormPageData struct {
	Item       *models.Item
	Categories []string
	Conditions []string
	Cities     []string
	IsEdit     bool
}

func ItemListHandler(w http.ResponseWriter, r *http.Request) {
	filter := &models.ItemFilter{
		Keyword:  getFormValue(r, "keyword"),
		Category: getFormValue(r, "category"),
		City:     getFormValue(r, "city"),
		Status:   getFormValue(r, "status"),
	}

	items := itemSvc.List(filter)

	statuses := map[string]string{
		models.ItemStatusOnSale:    "上架中",
		models.ItemStatusPending:   "待处理",
		models.ItemStatusCompleted: "已置换",
		models.ItemStatusOffline:   "已下架",
	}

	render(w, "index.html", &PageData{
		Title:      "货品列表",
		ActiveMenu: "home",
		Data: &ListPageData{
			Items:      items,
			Filter:     filter,
			Categories: models.Categories,
			Cities:     models.Cities,
			Statuses:   statuses,
		},
	})
}

func ItemDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	item, err := itemSvc.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	requests := exchangeSvc.ListByItem(id)

	render(w, "detail.html", &PageData{
		Title:      item.Title,
		ActiveMenu: "home",
		Data: &DetailPageData{
			Item:        item,
			Requests:    requests,
			CurrentUser: "模拟用户",
		},
	})
}

func ItemNewFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, "form.html", &PageData{
		Title:      "发布置换信息",
		ActiveMenu: "new",
		Data: &FormPageData{
			Categories: models.Categories,
			Conditions: models.Conditions,
			Cities:     models.Cities,
			IsEdit:     false,
		},
	})
}

func ItemCreateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item := &models.Item{
		Title:            getFormValue(r, "title"),
		Category:         getFormValue(r, "category"),
		Condition:        getFormValue(r, "condition"),
		City:             getFormValue(r, "city"),
		Owner:            getFormValue(r, "owner"),
		Description:      getFormValue(r, "description"),
		ExpectedExchange: getFormValue(r, "expected_exchange"),
	}

	if item.Owner == "" {
		item.Owner = "模拟用户"
	}

	_, err := itemSvc.Create(item)
	if err != nil {
		render(w, "form.html", &PageData{
			Title:      "发布置换信息",
			ActiveMenu: "new",
			Error:      err.Error(),
			Data: &FormPageData{
				Item:       item,
				Categories: models.Categories,
				Conditions: models.Conditions,
				Cities:     models.Cities,
				IsEdit:     false,
			},
		})
		return
	}

	http.Redirect(w, r, "/?success=created", http.StatusSeeOther)
}

func ItemEditFormHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	item, err := itemSvc.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	render(w, "form.html", &PageData{
		Title:      "编辑置换信息",
		ActiveMenu: "home",
		Data: &FormPageData{
			Item:       item,
			Categories: models.Categories,
			Conditions: models.Conditions,
			Cities:     models.Cities,
			IsEdit:     true,
		},
	})
}

func ItemUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := &models.Item{
		Title:            getFormValue(r, "title"),
		Category:         getFormValue(r, "category"),
		Condition:        getFormValue(r, "condition"),
		City:             getFormValue(r, "city"),
		Description:      getFormValue(r, "description"),
		ExpectedExchange: getFormValue(r, "expected_exchange"),
		Status:           getFormValue(r, "status"),
	}

	_, err := itemSvc.Update(id, updates)
	if err != nil {
		item, _ := itemSvc.Get(id)
		render(w, "form.html", &PageData{
			Title:      "编辑置换信息",
			ActiveMenu: "home",
			Error:      err.Error(),
			Data: &FormPageData{
				Item:       item,
				Categories: models.Categories,
				Conditions: models.Conditions,
				Cities:     models.Cities,
				IsEdit:     true,
			},
		})
		return
	}

	http.Redirect(w, r, "/items/"+id+"?success=updated", http.StatusSeeOther)
}

func ItemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	if err := itemSvc.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=deleted", http.StatusSeeOther)
}
