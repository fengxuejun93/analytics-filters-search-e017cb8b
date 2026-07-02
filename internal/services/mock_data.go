package services

import (
	"secondhand-exchange/internal/models"
)

func InitMockData() {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	mockItems := []*models.Item{
		{
			ID:               "item001",
			Title:            "MacBook Pro 2020 13寸",
			Category:         "数码电子",
			Condition:        "几乎全新",
			City:             "北京",
			Owner:            "数码达人",
			Description:      "2020款 MacBook Pro 13寸，M1芯片，16GB内存，512GB存储。电池循环次数89次，保养良好，无磕碰划痕。原包装盒和充电器齐全，购买凭证可查。",
			ExpectedExchange: "iPhone 13 Pro 或同价位 Windows 笔记本",
			Status:           models.ItemStatusOnSale,
			Images:           []string{"laptop"},
			CreatedAt:        now().AddDate(0, 0, -7),
			UpdatedAt:        now().AddDate(0, 0, -7),
		},
		{
			ID:               "item002",
			Title:            "索尼 A7M4 全画幅相机",
			Category:         "数码电子",
			Condition:        "轻微使用",
			City:             "上海",
			Owner:            "摄影爱好者",
			Description:      "索尼 A7M4 机身 + 24-70mm F4 镜头套装。快门次数约8000次，使用爱惜。屏幕贴了钢化膜，机身有轻微使用痕迹但不影响功能。",
			ExpectedExchange: "大疆 Mavic 3 无人机 或 等价摄影器材",
			Status:           models.ItemStatusOnSale,
			Images:           []string{"camera"},
			CreatedAt:        now().AddDate(0, 0, -5),
			UpdatedAt:        now().AddDate(0, 0, -5),
		},
		{
			ID:               "item003",
			Title:            "《人类简史》三部曲套装",
			Category:         "图书文具",
			Condition:        "几乎全新",
			City:             "广州",
			Owner:            "书虫小李",
			Description:      "包含《人类简史》《未来简史》《今日简史》三册，尤瓦尔·赫拉利著作。仅翻阅过一次，书角无折损，无笔记划线。",
			ExpectedExchange: "其他社科类书籍 或  Kindle 电子书阅读器",
			Status:           models.ItemStatusPending,
			Images:           []string{"books"},
			CreatedAt:        now().AddDate(0, 0, -10),
			UpdatedAt:        now().AddDate(0, 0, -3),
		},
		{
			ID:               "item004",
			Title:            "宜家双人沙发 灰色",
			Category:         "家居用品",
			Condition:        "轻微使用",
			City:             "深圳",
			Owner:            "搬家的小王",
			Description:      "宜家 KLIPPAN 双人沙发，灰色织物面料。使用两年，定期清洁保养，无污渍无破损。因搬家无法带走，诚意置换。",
			ExpectedExchange: "同价位餐桌椅套装 或 折叠床",
			Status:           models.ItemStatusOnSale,
			Images:           []string{"sofa"},
			CreatedAt:        now().AddDate(0, 0, -3),
			UpdatedAt:        now().AddDate(0, 0, -3),
		},
		{
			ID:               "item005",
			Title:            "Nike Air Max 运动鞋 42码",
			Category:         "服饰鞋帽",
			Condition:        "全新",
			City:             "杭州",
			Owner:            "潮鞋收藏家",
			Description:      "Nike Air Max 90 经典款，黑白配色，42码。全新未拆封，官方渠道购买，支持验货。因尺码不合适转出。",
			ExpectedExchange: "Adidas Yeezy 系列 或 其他同价位运动鞋",
			Status:           models.ItemStatusOnSale,
			Images:           []string{"shoes"},
			CreatedAt:        now().AddDate(0, 0, -2),
			UpdatedAt:        now().AddDate(0, 0, -2),
		},
		{
			ID:               "item006",
			Title:            "捷安特山地自行车",
			Category:         "运动户外",
			Condition:        "明显使用",
			City:             "成都",
			Owner:            "骑行爱好者",
			Description:      "捷安特 ATX 777 山地车，27速，碟刹。使用三年，有正常使用痕迹，各部件功能正常，定期保养维护。适合通勤或周末骑行。",
			ExpectedExchange: "公路自行车 或 健身器材套装",
			Status:           models.ItemStatusCompleted,
			Images:           []string{"bike"},
			CreatedAt:        now().AddDate(0, 0, -20),
			UpdatedAt:        now().AddDate(0, 0, -1),
		},
		{
			ID:               "item007",
			Title:            "戴森 V8 吸尘器",
			Category:         "家居用品",
			Condition:        "功能正常",
			City:             "武汉",
			Owner:            "居家小能手",
			Description:      "戴森 V8 Absolute 无线吸尘器，主机+5个吸头+充电底座。使用四年，电池续航约20分钟，吸力正常。外观有使用痕迹但功能完好。",
			ExpectedExchange: "扫地机器人 或 空气净化器",
			Status:           models.ItemStatusOnSale,
			Images:           []string{"vacuum"},
			CreatedAt:        now().AddDate(0, 0, -4),
			UpdatedAt:        now().AddDate(0, 0, -4),
		},
		{
			ID:               "item008",
			Title:            "iPad Pro 11寸 2021款",
			Category:         "数码电子",
			Condition:        "几乎全新",
			City:             "南京",
			Owner:            "学生小张",
			Description:      "iPad Pro 11寸 M1芯片，256GB WiFi版，深空灰色。带原装Apple Pencil 2代和妙控键盘。使用不到半年，因毕业准备换电脑转出。",
			ExpectedExchange: "16寸 MacBook Pro 或 台式机整机",
			Status:           models.ItemStatusOffline,
			Images:           []string{"ipad"},
			CreatedAt:        now().AddDate(0, 0, -15),
			UpdatedAt:        now().AddDate(0, 0, -8),
		},
	}

	for _, item := range mockItems {
		store.Items[item.ID] = item
	}

	mockRequests := []*models.ExchangeRequest{
		{
			ID:        "req001",
			ItemID:    "item003",
			Applicant: "爱读书的猫",
			OfferItem: "Kindle Paperwhite 5 电子书阅读器",
			Message:   "我有一台几乎全新的 Kindle Paperwhite 5，包装齐全，想换您的三部曲套装。可以面交验货~",
			Status:    models.ExchangeStatusPending,
			CreatedAt: now().AddDate(0, 0, -3),
			UpdatedAt: now().AddDate(0, 0, -3),
		},
		{
			ID:        "req002",
			ItemID:    "item006",
			Applicant: "健身达人",
			OfferItem: "全套健身器材（哑铃+杠铃+卧推凳）",
			Message:   "健身器材购入半年，使用次数不多，诚意置换。可以上门看货。",
			Status:    models.ExchangeStatusAccepted,
			CreatedAt: now().AddDate(0, 0, -10),
			UpdatedAt: now().AddDate(0, 0, -1),
		},
		{
			ID:        "req003",
			ItemID:    "item001",
			Applicant: "程序员小明",
			OfferItem: "ThinkPad X1 Carbon 2021",
			Message:   "ThinkPad X1 Carbon 2021，i7处理器，16GB内存，512GB固态。成色95新，想换您的 MacBook。",
			Status:    models.ExchangeStatusRejected,
			CreatedAt: now().AddDate(0, 0, -6),
			UpdatedAt: now().AddDate(0, 0, -5),
		},
	}

	for _, req := range mockRequests {
		store.ExchangeRequests[req.ID] = req
	}

	_ = store.saveLocked()
}
