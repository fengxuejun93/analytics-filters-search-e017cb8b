package models

import "time"

const (
	ItemStatusOnSale    = "on_sale"
	ItemStatusPending   = "pending"
	ItemStatusCompleted = "completed"
	ItemStatusOffline   = "offline"
)

const (
	ExchangeStatusPending  = "pending"
	ExchangeStatusAccepted = "accepted"
	ExchangeStatusRejected = "rejected"
	ExchangeStatusCanceled = "canceled"
)

var Categories = []string{"数码电子", "图书文具", "家居用品", "服饰鞋帽", "运动户外", "母婴用品", "其他"}

var Conditions = []string{"全新", "几乎全新", "轻微使用", "明显使用", "功能正常"}

var Cities = []string{"北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "南京", "西安", "重庆"}

var ItemStatusMap = map[string]string{
	ItemStatusOnSale:    "上架中",
	ItemStatusPending:   "待处理",
	ItemStatusCompleted: "已置换",
	ItemStatusOffline:   "已下架",
}

var ExchangeStatusMap = map[string]string{
	ExchangeStatusPending:  "待处理",
	ExchangeStatusAccepted: "已接受",
	ExchangeStatusRejected: "已拒绝",
	ExchangeStatusCanceled: "已取消",
}

type Item struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Category         string    `json:"category"`
	Condition        string    `json:"condition"`
	City             string    `json:"city"`
	Owner            string    `json:"owner"`
	Description      string    `json:"description"`
	ExpectedExchange string    `json:"expected_exchange"`
	Status           string    `json:"status"`
	Images           []string  `json:"images"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (i *Item) StatusText() string {
	if text, ok := ItemStatusMap[i.Status]; ok {
		return text
	}
	return i.Status
}

func (i *Item) StatusClass() string {
	switch i.Status {
	case ItemStatusOnSale:
		return "status-success"
	case ItemStatusPending:
		return "status-warning"
	case ItemStatusCompleted:
		return "status-info"
	case ItemStatusOffline:
		return "status-dark"
	default:
		return "status-default"
	}
}

type ExchangeRequest struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Applicant string    `json:"applicant"`
	OfferItem string    `json:"offer_item"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (e *ExchangeRequest) StatusText() string {
	if text, ok := ExchangeStatusMap[e.Status]; ok {
		return text
	}
	return e.Status
}

func (e *ExchangeRequest) StatusClass() string {
	switch e.Status {
	case ExchangeStatusPending:
		return "status-warning"
	case ExchangeStatusAccepted:
		return "status-success"
	case ExchangeStatusRejected:
		return "status-danger"
	case ExchangeStatusCanceled:
		return "status-dark"
	default:
		return "status-default"
	}
}

func (e *ExchangeRequest) CanAccept(itemOwner string) bool {
	return e.Status == ExchangeStatusPending && e.Applicant != itemOwner
}

func (e *ExchangeRequest) CanReject(itemOwner string) bool {
	return e.Status == ExchangeStatusPending && e.Applicant != itemOwner
}

func (e *ExchangeRequest) CanCancel(applicant string) bool {
	return e.Status == ExchangeStatusPending && e.Applicant == applicant
}

type Statistics struct {
	OnSaleCount    int `json:"on_sale_count"`
	PendingCount   int `json:"pending_count"`
	CompletedCount int `json:"completed_count"`
	OfflineCount   int `json:"offline_count"`
}

type ItemFilter struct {
	Keyword  string
	Category string
	City     string
	Status   string
}
