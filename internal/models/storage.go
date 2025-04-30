package models

import (
	"fmt"
	"math/rand"
	"time"
)

func (s *Storage) UpdateUsage() {
	s.UsedGB += rand.Float64() * 2
	s.LastAccessed = time.Now().Unix()
}
func (storage *Storage) GetId() string {
	return storage.ID
}

func (storage *Storage) GetUsage() float64 {
	return storage.UsedGB
}

func (storage *Storage) GetType() string {
	return "Storage"
}

func (s *Storage) String() string {
	return fmt.Sprintf("Storage[ID=%s, UsedGB=%.2f, CostPerGB=%.2f, PreviousCostPerGB=%.2f, LastAccessed=%d, Owner=%s]", s.ID, s.UsedGB, s.CostPerGB, s.PreviousCostPerGB, s.LastAccessed, s.Owner)
}
