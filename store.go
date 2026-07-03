package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Store 内存数据存储
type Store struct {
	mu            sync.RWMutex
	items         map[string]*Item
	applications  map[string]*Application
	history       map[string][]*StatusHistory
	itemSeq       int
	appSeq        int
	historySeq    int
}

func NewStore() *Store {
	s := &Store{
		items:        make(map[string]*Item),
		applications: make(map[string]*Application),
		history:      make(map[string][]*StatusHistory),
	}
	s.seedData()
	return s
}

func (s *Store) seedData() {
	items := []Item{
		{
			ID: "item-1", Title: "九成新 MacBook Pro 2021", Category: "数码电子", Condition: "九成新",
			City: "北京", ExpectedExchange: "iPad Pro 或显示器", Publisher: "张三",
			Status: ItemStatusListed, Description: "2021款 MacBook Pro 14寸，M1 Pro 芯片，16GB+512GB，使用一年半，无磕碰，电池健康度92%。希望能换一台 iPad Pro 或4K显示器。",
			ImageURL: "", CreatedAt: time.Now().Add(-72 * time.Hour), UpdatedAt: time.Now().Add(-72 * time.Hour),
		},
		{
			ID: "item-2", Title: "宜家 KALLAX 书架", Category: "家具家居", Condition: "八成新",
			City: "上海", ExpectedExchange: "落地灯或收纳柜", Publisher: "李四",
			Status: ItemStatusExchanged, Description: "宜家经典方格书架，白色2x4款式，搬家后空间不够用，有轻微使用痕迹。希望换落地灯或收纳柜。",
			ImageURL: "", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now().Add(-10 * time.Hour),
		},
		{
			ID: "item-3", Title: "Switch 健身环大冒险套装", Category: "数码电子", Condition: "九五新",
			City: "深圳", ExpectedExchange: "PS5 游戏或手柄", Publisher: "王五",
			Status: ItemStatusListed, Description: "健身环大冒险全套含环和腿带，Switch 主机不包含，使用不到三个月，几乎全新。希望换PS5游戏或手柄。",
			ImageURL: "", CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID: "item-4", Title: "全汉骑行自行车", Category: "运动户外", Condition: "七成新",
			City: "广州", ExpectedExchange: "跑步机或椭圆机", Publisher: "赵六",
			Status: ItemStatusListed, Description: "全汉公路自行车，铝合金车架，变速正常，有些锈迹但骑起来没问题。想换室内运动器材。",
			ImageURL: "", CreatedAt: time.Now().Add(-12 * time.Hour), UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID: "item-5", Title: "《百年孤独》等十本经典文学", Category: "图书文具", Condition: "八成新",
			City: "成都", ExpectedExchange: "科幻小说或漫画合集", Publisher: "孙七",
			Status: ItemStatusListed, Description: "包含百年孤独、1984、美丽新世界等十本经典文学名著，均有阅读痕迹但保存良好。希望换科幻小说或漫画。",
			ImageURL: "", CreatedAt: time.Now().Add(-6 * time.Hour), UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			ID: "item-6", Title: "Sony WH-1000XM4 降噪耳机", Category: "数码电子", Condition: "九成新",
			City: "北京", ExpectedExchange: "AirPods Pro 或音箱", Publisher: "周八",
			Status: ItemStatusDelisted, Description: "Sony 头戴式降噪耳机，黑色，降噪效果极佳，配件齐全。已与别人达成置换，下架处理。",
			ImageURL: "", CreatedAt: time.Now().Add(-120 * time.Hour), UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for i := range items {
		s.items[items[i].ID] = &items[i]
	}
	s.itemSeq = 6

	apps := []Application{
		{
			ID: "app-1", ItemID: "item-1", Applicant: "陈一", OfferItem: "iPad Pro 2022 11寸 128GB", Message: "我有一台 iPad Pro 2022 11寸 128GB，感兴趣吗？",
			Status: AppStatusPending, CreatedAt: time.Now().Add(-36 * time.Hour), UpdatedAt: time.Now().Add(-36 * time.Hour),
		},
		{
			ID: "app-2", ItemID: "item-1", Applicant: "吴二", OfferItem: "LG 4K 显示器 27寸", Message: "有一台 LG 4K 显示器 27寸，可以加一些差价。",
			Status: AppStatusPending, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID: "app-3", ItemID: "item-2", Applicant: "郑三", OfferItem: "飞利浦落地灯 九成新", Message: "我有一款飞利浦落地灯，九成新，可以交换。",
			Status: AppStatusAccepted, CreatedAt: time.Now().Add(-40 * time.Hour), UpdatedAt: time.Now().Add(-10 * time.Hour),
		},
		{
			ID: "app-4", ItemID: "item-4", Applicant: "冯四", OfferItem: "小型跑步机", Message: "有一台小型跑步机，用了一年，可以换吗？",
			Status: AppStatusRejected, CreatedAt: time.Now().Add(-8 * time.Hour), UpdatedAt: time.Now().Add(-4 * time.Hour),
		},
		{
			ID: "app-5", ItemID: "item-6", Applicant: "何五", OfferItem: "AirPods Pro 2代", Message: "AirPods Pro 2代可以换。",
			Status: AppStatusAccepted, CreatedAt: time.Now().Add(-96 * time.Hour), UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for i := range apps {
		s.applications[apps[i].ID] = &apps[i]
	}
	s.appSeq = 5

	// 种子历史记录
	s.historySeq = 10
	s.history["item-1"] = []*StatusHistory{
		{ID: "h-1", ItemID: "item-1", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "张三", CreatedAt: time.Now().Add(-72 * time.Hour)},
		{ID: "h-1a", ItemID: "item-1", FromStatus: "listed", ToStatus: "listed", Reason: "陈一 发起置换申请（提供: iPad Pro 2022 11寸 128GB）", Operator: "陈一", CreatedAt: time.Now().Add(-36 * time.Hour)},
		{ID: "h-1b", ItemID: "item-1", FromStatus: "listed", ToStatus: "listed", Reason: "吴二 发起置换申请（提供: LG 4K 显示器 27寸）", Operator: "吴二", CreatedAt: time.Now().Add(-24 * time.Hour)},
	}
	s.history["item-2"] = []*StatusHistory{
		{ID: "h-2", ItemID: "item-2", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "李四", CreatedAt: time.Now().Add(-48 * time.Hour)},
		{ID: "h-3", ItemID: "item-2", FromStatus: "listed", ToStatus: "exchanged", Reason: "接受 郑三 的置换申请", Operator: "李四", CreatedAt: time.Now().Add(-10 * time.Hour)},
	}
	s.history["item-3"] = []*StatusHistory{
		{ID: "h-4", ItemID: "item-3", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "王五", CreatedAt: time.Now().Add(-24 * time.Hour)},
	}
	s.history["item-4"] = []*StatusHistory{
		{ID: "h-5", ItemID: "item-4", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "赵六", CreatedAt: time.Now().Add(-12 * time.Hour)},
		{ID: "h-5a", ItemID: "item-4", FromStatus: "listed", ToStatus: "listed", Reason: "冯四 发起置换申请（提供: 小型跑步机）", Operator: "冯四", CreatedAt: time.Now().Add(-8 * time.Hour)},
		{ID: "h-5b", ItemID: "item-4", FromStatus: "listed", ToStatus: "listed", Reason: "拒绝 冯四 的置换申请", Operator: "赵六", CreatedAt: time.Now().Add(-4 * time.Hour)},
	}
	s.history["item-5"] = []*StatusHistory{
		{ID: "h-6", ItemID: "item-5", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "孙七", CreatedAt: time.Now().Add(-6 * time.Hour)},
	}
	s.history["item-6"] = []*StatusHistory{
		{ID: "h-7", ItemID: "item-6", FromStatus: "", ToStatus: "listed", Reason: "发布上架", Operator: "周八", CreatedAt: time.Now().Add(-120 * time.Hour)},
		{ID: "h-8", ItemID: "item-6", FromStatus: "listed", ToStatus: "exchanged", Reason: "接受 何五 的置换申请", Operator: "周八", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "h-9", ItemID: "item-6", FromStatus: "exchanged", ToStatus: "delisted", Reason: "主动下架", Operator: "周八", CreatedAt: time.Now().Add(-1 * time.Hour)},
	}
}

// ========== 货品操作 ==========

// addHistory 记录状态变更历史（调用方需持有写锁）
func (s *Store) addHistory(itemID, fromStatus, toStatus, reason, operator string) {
	s.historySeq++
	h := &StatusHistory{
		ID:         fmt.Sprintf("h-%d", s.historySeq),
		ItemID:     itemID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Reason:     reason,
		Operator:   operator,
		CreatedAt:  time.Now(),
	}
	s.history[itemID] = append(s.history[itemID], h)
}

// GetStatusHistory 获取货品状态历史
func (s *Store) GetStatusHistory(itemID string) []*StatusHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history[itemID]
}

// GetItemDetail 获取聚合详情
func (s *Store) GetItemDetail(itemID string) (*ItemDetailResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[itemID]
	if !ok {
		return nil, false
	}

	var apps []*Application
	for _, app := range s.applications {
		if app.ItemID == itemID {
			apps = append(apps, app)
		}
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].CreatedAt.After(apps[j].CreatedAt)
	})

	pendingCount := 0
	for _, a := range apps {
		if a.Status == AppStatusPending {
			pendingCount++
		}
	}

	history := s.history[itemID]

	return &ItemDetailResponse{
		Item:          item,
		Applications:  apps,
		StatusHistory: history,
		PendingCount:  pendingCount,
	}, true
}

func (s *Store) ListItems(filter FilterParams) []*Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Item
	for _, item := range s.items {
		if filter.Keyword != "" {
			kw := strings.ToLower(filter.Keyword)
			if !strings.Contains(strings.ToLower(item.Title), kw) &&
				!strings.Contains(strings.ToLower(item.Description), kw) &&
				!strings.Contains(strings.ToLower(item.ExpectedExchange), kw) {
				continue
			}
		}
		if filter.Category != "" && item.Category != filter.Category {
			continue
		}
		if filter.City != "" && item.City != filter.City {
			continue
		}
		if filter.Status != "" && string(item.Status) != filter.Status {
			continue
		}
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

func (s *Store) GetItem(id string) (*Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	return item, true
}

func (s *Store) CreateItem(req ItemRequest) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.itemSeq++
	now := time.Now()
	item := &Item{
		ID:               fmt.Sprintf("item-%d", s.itemSeq),
		Title:            req.Title,
		Category:         req.Category,
		Condition:        req.Condition,
		City:             req.City,
		ExpectedExchange: req.ExpectedExchange,
		Publisher:        req.Publisher,
		Status:           ItemStatusListed,
		Description:      req.Description,
		ImageURL:         req.ImageURL,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.items[item.ID] = item
	s.addHistory(item.ID, "", "listed", "发布上架", req.Publisher)
	return item
}

func (s *Store) UpdateItem(id string, req ItemRequest) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	item.Title = req.Title
	item.Category = req.Category
	item.Condition = req.Condition
	item.City = req.City
	item.ExpectedExchange = req.ExpectedExchange
	item.Publisher = req.Publisher
	item.Description = req.Description
	item.ImageURL = req.ImageURL
	item.UpdatedAt = time.Now()
	s.addHistory(id, string(item.Status), string(item.Status), "编辑货品信息", req.Publisher)
	return item, true
}

func (s *Store) UpdateItemStatus(id string, status ItemStatus) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return nil, false
	}
	oldStatus := string(item.Status)
	item.Status = status
	item.UpdatedAt = time.Now()
	reasonMap := map[string]string{
		"listed":   "重新上架/恢复上架",
		"delisted": "主动下架",
		"exchanged": "已达成置换",
	}
	reason := reasonMap[string(status)]
	if reason == "" {
		reason = fmt.Sprintf("状态变更为 %s", status)
	}
	s.addHistory(id, oldStatus, string(status), reason, item.Publisher)
	return item, true
}

// ========== 置换申请操作 ==========

func (s *Store) ListApplications(itemID string) []*Application {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Application
	for _, app := range s.applications {
		if app.ItemID == itemID {
			result = append(result, app)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

func (s *Store) GetApplication(id string) (*Application, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.applications[id]
	if !ok {
		return nil, false
	}
	return app, true
}

func (s *Store) CreateApplication(itemID string, req ApplicationRequest) (*Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok {
		return nil, fmt.Errorf("货品不存在")
	}
	if item.Status != ItemStatusListed {
		return nil, fmt.Errorf("该货品当前状态不允许发起申请")
	}

	s.appSeq++
	now := time.Now()
	app := &Application{
		ID:        fmt.Sprintf("app-%d", s.appSeq),
		ItemID:    itemID,
		Applicant: req.Applicant,
		OfferItem: req.OfferItem,
		Message:   req.Message,
		Status:    AppStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.applications[app.ID] = app

	offerDesc := req.OfferItem
	if offerDesc == "" {
		offerDesc = "未指定"
	}
	s.addHistory(itemID, string(item.Status), string(item.Status),
		fmt.Sprintf("%s 发起置换申请（提供: %s）", req.Applicant, offerDesc), req.Applicant)

	return app, nil
}

func (s *Store) HandleApplication(appID string, action string) (*Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, ok := s.applications[appID]
	if !ok {
		return nil, fmt.Errorf("申请不存在")
	}

	switch app.Status {
	case AppStatusPending:
		switch action {
		case "accept":
			app.Status = AppStatusAccepted
			// 接受申请后，将货品标记为已置换
			if item, ok := s.items[app.ItemID]; ok {
				oldSt := string(item.Status)
				item.Status = ItemStatusExchanged
				item.UpdatedAt = time.Now()
				s.addHistory(app.ItemID, oldSt, "exchanged", fmt.Sprintf("接受 %s 的置换申请", app.Applicant), item.Publisher)
			}
			// 拒绝该货品的其他待处理申请
			for _, other := range s.applications {
				if other.ItemID == app.ItemID && other.ID != app.ID && other.Status == AppStatusPending {
					other.Status = AppStatusRejected
					other.UpdatedAt = time.Now()
					s.addHistory(app.ItemID, "listed", "listed", fmt.Sprintf("自动拒绝 %s 的申请（已接受其他申请）", other.Applicant), "系统")
				}
			}
		case "reject":
			app.Status = AppStatusRejected
			if item, ok := s.items[app.ItemID]; ok {
				s.addHistory(app.ItemID, string(item.Status), string(item.Status), fmt.Sprintf("拒绝 %s 的置换申请", app.Applicant), item.Publisher)
			}
		case "cancel":
			app.Status = AppStatusCancelled
			if item, ok := s.items[app.ItemID]; ok {
				s.addHistory(app.ItemID, string(item.Status), string(item.Status), fmt.Sprintf("%s 取消置换申请", app.Applicant), app.Applicant)
			}
		default:
			return nil, fmt.Errorf("不支持的操作: %s", action)
		}
	case AppStatusAccepted:
		if action == "cancel" {
			app.Status = AppStatusCancelled
			// 取消已接受的申请，恢复货品为上架中
			if item, ok := s.items[app.ItemID]; ok {
				oldSt := string(item.Status)
				item.Status = ItemStatusListed
				item.UpdatedAt = time.Now()
				s.addHistory(app.ItemID, oldSt, "listed", fmt.Sprintf("取消 %s 的置换申请，恢复上架", app.Applicant), item.Publisher)
			}
		} else {
			return nil, fmt.Errorf("已接受的申请只能取消，不能再次接受或拒绝")
		}
	case AppStatusRejected:
		return nil, fmt.Errorf("已拒绝的申请无法再操作")
	case AppStatusCancelled:
		return nil, fmt.Errorf("已取消的申请无法再操作")
	}

	app.UpdatedAt = time.Now()
	return app, nil
}

// ========== 统计 ==========

func (s *Store) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Stats{TotalItems: len(s.items), TotalApps: len(s.applications)}
	for _, item := range s.items {
		switch item.Status {
		case ItemStatusListed:
			stats.ListedCount++
		case ItemStatusExchanged:
			stats.ExchangedCount++
		case ItemStatusDelisted:
			stats.DelistedCount++
		}
	}
	for _, app := range s.applications {
		if app.Status == AppStatusPending {
			stats.PendingCount++
		}
	}
	return stats
}

// 获取所有品类
func (s *Store) GetCategories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var result []string
	for _, item := range s.items {
		if !seen[item.Category] {
			seen[item.Category] = true
			result = append(result, item.Category)
		}
	}
	sort.Strings(result)
	return result
}

// 获取所有城市
func (s *Store) GetCities() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var result []string
	for _, item := range s.items {
		if !seen[item.City] {
			seen[item.City] = true
			result = append(result, item.City)
		}
	}
	sort.Strings(result)
	return result
}
