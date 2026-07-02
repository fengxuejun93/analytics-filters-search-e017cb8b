package services

import (
	"secondhand-exchange/internal/models"
)

type StatsService struct{}

func NewStatsService() *StatsService {
	return &StatsService{}
}

func (s *StatsService) GetStatistics() *models.Statistics {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	stats := &models.Statistics{}

	for _, item := range store.Items {
		switch item.Status {
		case models.ItemStatusOnSale:
			stats.OnSaleCount++
		case models.ItemStatusPending:
			stats.PendingCount++
		case models.ItemStatusCompleted:
			stats.CompletedCount++
		case models.ItemStatusOffline:
			stats.OfflineCount++
		}
	}

	return stats
}

func (s *StatsService) GetPendingExchangeCount() int {
	store := GetStore()
	store.mu.RLock()
	defer store.mu.RUnlock()

	count := 0
	for _, req := range store.ExchangeRequests {
		if req.Status == models.ExchangeStatusPending {
			count++
		}
	}
	return count
}
