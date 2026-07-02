package services

import (
	"errors"

	"secondhand-exchange/internal/models"
)

type ExchangeService struct {
	itemService *ItemService
}

func NewExchangeService() *ExchangeService {
	return &ExchangeService{
		itemService: NewItemService(),
	}
}

func (s *ExchangeService) ListByItem(itemID string) []*models.ExchangeRequest {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	requests := make([]*models.ExchangeRequest, 0)
	for _, req := range store.ExchangeRequests {
		if req.ItemID == itemID {
			requests = append(requests, req)
		}
	}

	return requests
}

func (s *ExchangeService) Get(id string) (*models.ExchangeRequest, error) {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	req, exists := store.ExchangeRequests[id]
	if !exists {
		return nil, errors.New("exchange request not found")
	}
	return req, nil
}

func (s *ExchangeService) Apply(itemID, applicant, offerItem, message string) (*models.ExchangeRequest, error) {
	if itemID == "" {
		return nil, errors.New("item id is required")
	}
	if applicant == "" {
		return nil, errors.New("applicant is required")
	}
	if offerItem == "" {
		return nil, errors.New("offer item is required")
	}

	item, err := s.itemService.Get(itemID)
	if err != nil {
		return nil, err
	}

	if item.Status != models.ItemStatusOnSale && item.Status != models.ItemStatusPending {
		return nil, errors.New("item is not available for exchange")
	}

	if item.Owner == applicant {
		return nil, errors.New("cannot apply for your own item")
	}

	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	req := &models.ExchangeRequest{
		ID:        generateID(),
		ItemID:    itemID,
		Applicant: applicant,
		OfferItem: offerItem,
		Message:   message,
		Status:    models.ExchangeStatusPending,
		CreatedAt: now(),
		UpdatedAt: now(),
	}

	store.ExchangeRequests[req.ID] = req

	if item.Status == models.ItemStatusOnSale {
		item.Status = models.ItemStatusPending
		item.UpdatedAt = now()
	}

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *ExchangeService) Accept(id, operator string) (*models.ExchangeRequest, error) {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	req, exists := store.ExchangeRequests[id]
	if !exists {
		return nil, errors.New("exchange request not found")
	}

	if req.Status != models.ExchangeStatusPending {
		return nil, errors.New("exchange request is not pending")
	}

	item, exists := store.Items[req.ItemID]
	if !exists {
		return nil, errors.New("item not found")
	}

	if item.Owner != operator {
		return nil, errors.New("only item owner can accept")
	}

	req.Status = models.ExchangeStatusAccepted
	req.UpdatedAt = now()

	item.Status = models.ItemStatusCompleted
	item.UpdatedAt = now()

	for _, otherReq := range store.ExchangeRequests {
		if otherReq.ItemID == req.ItemID && otherReq.ID != req.ID && otherReq.Status == models.ExchangeStatusPending {
			otherReq.Status = models.ExchangeStatusRejected
			otherReq.UpdatedAt = now()
		}
	}

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *ExchangeService) Reject(id, operator string) (*models.ExchangeRequest, error) {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	req, exists := store.ExchangeRequests[id]
	if !exists {
		return nil, errors.New("exchange request not found")
	}

	if req.Status != models.ExchangeStatusPending {
		return nil, errors.New("exchange request is not pending")
	}

	item, exists := store.Items[req.ItemID]
	if !exists {
		return nil, errors.New("item not found")
	}

	if item.Owner != operator {
		return nil, errors.New("only item owner can reject")
	}

	req.Status = models.ExchangeStatusRejected
	req.UpdatedAt = now()

	hasPending := false
	for _, otherReq := range store.ExchangeRequests {
		if otherReq.ItemID == req.ItemID && otherReq.ID != req.ID && otherReq.Status == models.ExchangeStatusPending {
			hasPending = true
			break
		}
	}

	if !hasPending && item.Status == models.ItemStatusPending {
		item.Status = models.ItemStatusOnSale
		item.UpdatedAt = now()
	}

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *ExchangeService) Cancel(id, operator string) (*models.ExchangeRequest, error) {
	store := GetStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	req, exists := store.ExchangeRequests[id]
	if !exists {
		return nil, errors.New("exchange request not found")
	}

	if req.Status != models.ExchangeStatusPending {
		return nil, errors.New("exchange request is not pending")
	}

	if req.Applicant != operator {
		return nil, errors.New("only applicant can cancel")
	}

	req.Status = models.ExchangeStatusCanceled
	req.UpdatedAt = now()

	item, exists := store.Items[req.ItemID]
	if exists {
		hasPending := false
		for _, otherReq := range store.ExchangeRequests {
			if otherReq.ItemID == req.ItemID && otherReq.ID != req.ID && otherReq.Status == models.ExchangeStatusPending {
				hasPending = true
				break
			}
		}

		if !hasPending && item.Status == models.ItemStatusPending {
			item.Status = models.ItemStatusOnSale
			item.UpdatedAt = now()
		}
	}

	if err := store.saveLocked(); err != nil {
		return nil, err
	}

	return req, nil
}
