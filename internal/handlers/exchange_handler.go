package handlers

import (
	"net/http"
	"net/url"
)

func ExchangeApplyHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itemID := getFormValue(r, "item_id")
	applicant := getFormValue(r, "applicant")
	offerItem := getFormValue(r, "offer_item")
	message := getFormValue(r, "message")

	if applicant == "" {
		applicant = "模拟用户"
	}

	_, err := exchangeSvc.Apply(itemID, applicant, offerItem, message)
	if err != nil {
		query := url.Values{}
		query.Set("error", url.QueryEscape(err.Error()))
		http.Redirect(w, r, "/items/"+itemID+"?"+query.Encode(), http.StatusSeeOther)
		return
	}

	query := url.Values{}
	query.Set("success", "申请已提交")
	http.Redirect(w, r, "/items/"+itemID+"?"+query.Encode(), http.StatusSeeOther)
}

func ExchangeAcceptHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	operator := "模拟用户"

	req, err := exchangeSvc.Accept(id, operator)
	if err != nil {
		referer := r.Header.Get("Referer")
		if referer == "" {
			referer = "/"
		}
		query := url.Values{}
		query.Set("error", url.QueryEscape(err.Error()))
		if u, err := url.Parse(referer); err == nil {
			u.RawQuery = query.Encode()
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, referer+"?"+query.Encode(), http.StatusSeeOther)
		}
		return
	}

	query := url.Values{}
	query.Set("success", "已接受置换申请")
	http.Redirect(w, r, "/items/"+req.ItemID+"?"+query.Encode(), http.StatusSeeOther)
}

func ExchangeRejectHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	operator := "模拟用户"

	req, err := exchangeSvc.Reject(id, operator)
	if err != nil {
		referer := r.Header.Get("Referer")
		if referer == "" {
			referer = "/"
		}
		query := url.Values{}
		query.Set("error", url.QueryEscape(err.Error()))
		if u, err := url.Parse(referer); err == nil {
			u.RawQuery = query.Encode()
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, referer+"?"+query.Encode(), http.StatusSeeOther)
		}
		return
	}

	query := url.Values{}
	query.Set("success", "已拒绝置换申请")
	http.Redirect(w, r, "/items/"+req.ItemID+"?"+query.Encode(), http.StatusSeeOther)
}

func ExchangeCancelHandler(w http.ResponseWriter, r *http.Request) {
	id := getPathValue(r, "id")
	operator := "模拟用户"

	req, err := exchangeSvc.Cancel(id, operator)
	if err != nil {
		referer := r.Header.Get("Referer")
		if referer == "" {
			referer = "/"
		}
		query := url.Values{}
		query.Set("error", url.QueryEscape(err.Error()))
		if u, err := url.Parse(referer); err == nil {
			u.RawQuery = query.Encode()
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, referer+"?"+query.Encode(), http.StatusSeeOther)
		}
		return
	}

	query := url.Values{}
	query.Set("success", "已取消置换申请")
	http.Redirect(w, r, "/items/"+req.ItemID+"?"+query.Encode(), http.StatusSeeOther)
}
