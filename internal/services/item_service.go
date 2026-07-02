package services

import (
	"errors"
	"strings"

	"secondhand-exchange/internal/models"
)

type ItemService struct{}

func NewItemService() *ItemService {
	return &ItemService{}
}

func (s *ItemService) List(filter *models.ItemFilter) []*models.Item {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	items := make([]*models.Item, 0, len(store.Items))

	for _, item := range store.Items {
		if filter != nil {
			if filter.Keyword != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(filter.Keyword)) &&
				!strings.Contains(strings.ToLower(item.Description), strings.ToLower(filter.Keyword)) {
				continue
			}
			if filter.Category != "" && item.Category != filter.Category {
				continue
			}
			if filter.City != "" && item.City != filter.City {
				continue
			}
			if filter.Status != "" && item.Status != filter.Status {
				continue
			}
		}
		items = append(items, item)
	}

	return items
}

func (s *ItemService) Get(id string) (*models.Item, error) {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	item, exists := store.Items[id]
	if !exists {
		return nil, errors.New("item not found")
	}
	return item, nil
}

func (s *ItemService) Create(item *models.Item) (*models.Item, error) {
	if item.Title == "" {
		return nil, errors.New("title is required")
	}
	if item.Category == "" {
		return nil, errors.New("category is required")
	}
	if item.City == "" {
		return nil, errors.New("city is required")
	}
	if item.Owner == "" {
		return nil, errors.New("owner is required")
	}

	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	item.ID = generateID()
	item.Status = models.ItemStatusOnSale
	item.CreatedAt = now()
	item.UpdatedAt = now()
	if item.Images == nil {
		item.Images = []string{}
	}

	store.Items[item.ID] = item

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) Update(id string, updates *models.Item) (*models.Item, error) {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	item, exists := store.Items[id]
	if !exists {
		return nil, errors.New("item not found")
	}

	if updates.Title != "" {
		item.Title = updates.Title
	}
	if updates.Category != "" {
		item.Category = updates.Category
	}
	if updates.Condition != "" {
		item.Condition = updates.Condition
	}
	if updates.City != "" {
		item.City = updates.City
	}
	if updates.Description != "" {
		item.Description = updates.Description
	}
	if updates.ExpectedExchange != "" {
		item.ExpectedExchange = updates.ExpectedExchange
	}
	if updates.Status != "" {
		item.Status = updates.Status
	}
	if updates.Images != nil {
		item.Images = updates.Images
	}

	item.UpdatedAt = now()

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) Delete(id string) error {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.Items[id]; !exists {
		return errors.New("item not found")
	}

	item := store.Items[id]
	item.Status = models.ItemStatusOffline
	item.UpdatedAt = now()

	for _, req := range store.ExchangeRequests {
		if req.ItemID == id && req.Status == models.ExchangeStatusPending {
			req.Status = models.ExchangeStatusCanceled
			req.UpdatedAt = now()
		}
	}

	if err := store.saveLocked(); err != nil {
		return err
	}

	return nil
}

func (s *ItemService) UpdateStatus(id string, status string) error {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	item, exists := store.Items[id]
	if !exists {
		return errors.New("item not found")
	}

	item.Status = status
	item.UpdatedAt = now()

	return store.saveLocked()
}

func (s *ItemService) HasPendingExchange(itemID string) bool {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, req := range store.ExchangeRequests {
		if req.ItemID == itemID && req.Status == models.ExchangeStatusPending {
			return true
		}
	}
	return false
}
