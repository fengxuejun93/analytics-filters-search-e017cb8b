package main

import "time"

// 货品状态
type ItemStatus string

const (
	ItemStatusListed   ItemStatus = "listed"   // 上架中
	ItemStatusDelisted ItemStatus = "delisted" // 已下架
	ItemStatusExchanged ItemStatus = "exchanged" // 已置换
)

// 置换申请状态
type ApplicationStatus string

const (
	AppStatusPending  ApplicationStatus = "pending"  // 待处理
	AppStatusAccepted ApplicationStatus = "accepted" // 已接受
	AppStatusRejected ApplicationStatus = "rejected" // 已拒绝
	AppStatusCancelled ApplicationStatus = "cancelled" // 已取消
)

// 状态历史记录
type StatusHistory struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	FromStatus string   `json:"from_status"`
	ToStatus   string   `json:"to_status"`
	Reason    string    `json:"reason"`
	Operator  string    `json:"operator"`
	CreatedAt time.Time `json:"created_at"`
}

// 货品模型
type Item struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Category         string     `json:"category"`
	Condition        string     `json:"condition"`
	City             string     `json:"city"`
	ExpectedExchange string     `json:"expected_exchange"`
	Publisher        string     `json:"publisher"`
	Status           ItemStatus `json:"status"`
	Description      string     `json:"description"`
	ImageURL         string     `json:"image_url"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// 置换申请模型
type Application struct {
	ID        string            `json:"id"`
	ItemID    string            `json:"item_id"`
	Applicant string            `json:"applicant"`
	OfferItem string            `json:"offer_item"`
	Message   string            `json:"message"`
	Status    ApplicationStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// 统计数据
type Stats struct {
	ListedCount    int `json:"listed_count"`
	PendingCount   int `json:"pending_count"`
	ExchangedCount int `json:"exchanged_count"`
	DelistedCount  int `json:"delisted_count"`
	TotalItems     int `json:"total_items"`
	TotalApps      int `json:"total_apps"`
}

// 创建/编辑货品请求
type ItemRequest struct {
	Title            string `json:"title"`
	Category         string `json:"category"`
	Condition        string `json:"condition"`
	City             string `json:"city"`
	ExpectedExchange string `json:"expected_exchange"`
	Publisher        string `json:"publisher"`
	Description      string `json:"description"`
	ImageURL         string `json:"image_url"`
}

// 创建置换申请请求
type ApplicationRequest struct {
	Applicant string `json:"applicant"`
	OfferItem string `json:"offer_item"`
	Message   string `json:"message"`
}

// 处理置换申请请求
type ApplicationActionRequest struct {
	Action string `json:"action"` // accept, reject, cancel
}

// 聚合详情响应
type ItemDetailResponse struct {
	Item           *Item            `json:"item"`
	Applications   []*Application   `json:"applications"`
	StatusHistory  []*StatusHistory `json:"status_history"`
	PendingCount   int              `json:"pending_count"`
}

// 筛选参数
type FilterParams struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	City     string `json:"city"`
	Status   string `json:"status"`
}
