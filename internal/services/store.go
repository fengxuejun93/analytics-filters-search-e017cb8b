package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"secondhand-exchange/internal/models"
)

type Store struct {
	mu               sync.RWMutex
	Items            map[string]*models.Item
	ExchangeRequests map[string]*models.ExchangeRequest
	dataFile         string
}

var store *Store

func InitStore(dataDir string) error {
	dataFile := filepath.Join(dataDir, "mock_data.json")
	store = &Store{
		Items:            make(map[string]*models.Item),
		ExchangeRequests: make(map[string]*models.ExchangeRequest),
		dataFile:         dataFile,
	}

	if err := store.load(); err != nil {
		return err
	}

	if len(store.Items) == 0 {
		InitMockData()
	}

	return nil
}

func GetStore() *Store {
	return store
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.dataFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var temp struct {
		Items            []*models.Item            `json:"items"`
		ExchangeRequests []*models.ExchangeRequest `json:"exchange_requests"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	for _, item := range temp.Items {
		s.Items[item.ID] = item
	}

	for _, req := range temp.ExchangeRequests {
		s.ExchangeRequests[req.ID] = req
	}

	return nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	temp := struct {
		Items            []*models.Item            `json:"items"`
		ExchangeRequests []*models.ExchangeRequest `json:"exchange_requests"`
	}{
		Items:            make([]*models.Item, 0, len(s.Items)),
		ExchangeRequests: make([]*models.ExchangeRequest, 0, len(s.ExchangeRequests)),
	}

	for _, item := range s.Items {
		temp.Items = append(temp.Items, item)
	}

	for _, req := range s.ExchangeRequests {
		temp.ExchangeRequests = append(temp.ExchangeRequests, req)
	}

	data, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.dataFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

func now() time.Time {
	return time.Now().Truncate(time.Second)
}
