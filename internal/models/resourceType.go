package models

import "time"

type CloudResource interface {
	UpdateUsage()
	GetId() string
	GetUsage() float64
	GetType() string
}

type VM struct {
	ID               string
	CPUUsage         float64
	CostPerHour      float64
	PreviousCostPerHour float64 
	Owner            string
	LastActive       int64 
}

type Storage struct {
	ID              string
	UsedGB          float64
	CostPerGB       float64
	PreviousCostPerGB float64 
	LastAccessed    int64
	Owner           string
}

type Lambda struct {
	ID             string
	Invocations    int
	Errors         int
	CostPerMillion float64
	Owner          string
	LastModified   int64
}

type ELB struct {
	ID           string
	RequestCount int
	HealthyHosts int
	CostPerHour  float64
	Owner        string
	LastChecked  int64
}

type S3 struct {
	ID          string
	UsedGB      float64
	ObjectCount int
	CostPerGB   float64
	Owner       string
	LastAccessed int64
}

type DynamoDB struct {
	ID           string
	ReadCapacity int
	WriteCapacity int
	ItemCount    int
	CostPerHr    float64
	Owner        string
	LastUpdated  int64
}

func (d *DynamoDB) UpdateUsage() {
	d.ItemCount += 100 + int(time.Now().Unix()%30)
	d.LastUpdated = time.Now().Unix()
}

func (d *DynamoDB) GetId() string {
	return d.ID
}

func (d *DynamoDB) GetUsage() float64 {
	return float64(d.ItemCount)
}

func (d *DynamoDB) GetType() string {
	return "DynamoDB"
}

func (s *S3) UpdateUsage() {
	s.UsedGB += 1.0 + float64(time.Now().Unix()%10)/10.0
	s.ObjectCount += 100 + int(time.Now().Unix()%20)
	s.LastAccessed = time.Now().Unix()
}

func (s *S3) GetId() string {
	return s.ID
}

func (s *S3) GetUsage() float64 {
	return s.UsedGB
}

func (s *S3) GetType() string {
	return "S3"
}

func (e *ELB) UpdateUsage() {
	e.RequestCount += 1000 + int(time.Now().Unix()%100)
	e.HealthyHosts = 2 + int(time.Now().Unix()%3)
	e.LastChecked = time.Now().Unix()
}

func (e *ELB) GetId() string {
	return e.ID
}

func (e *ELB) GetUsage() float64 {
	return float64(e.RequestCount)
}

func (e *ELB) GetType() string {
	return "ELB"
}

func (l *Lambda) UpdateUsage() {
	l.Invocations += 100 + int(time.Now().Unix()%50)
	l.Errors += int(time.Now().Unix() % 3)
	l.LastModified = time.Now().Unix()
}

func (l *Lambda) GetId() string {
	return l.ID
}

func (l *Lambda) GetUsage() float64 {
	return float64(l.Invocations)
}

func (l *Lambda) GetType() string {
	return "Lambda"
}

type Database struct {
	ID              string
	Connections     int
	CPUUsage        float64
	CostPerHr       float64
	PreviousCostPerHr float64 
	Owner           string
}
