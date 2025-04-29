package models

type CloudResource interface {
	UpdateUsage()
	GetId() string
	GetUsage() float64
	GetType() string
}

type VM struct {
	ID          string
	CPUUsage    float64
	CostPerHour float64
}

type Storage struct {
	ID           string
	UsedGB       float64
	CostPerGB    float64
	LastAccessed int64
}

type Database struct {
	ID          string
	Connections int
	CPUUsage    float64
	CostPerHr   float64
}
