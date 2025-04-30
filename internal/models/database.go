package models

import (
	"fmt"
	"math/rand"
)

func (d *Database) UpdateUsage() {
	d.Connections = rand.Intn(200)
	d.CPUUsage = rand.Float64() * 80
}

func (db *Database) GetId() string {
	return db.ID
}

func (db *Database) GetUsage() float64 {
	return float64(db.Connections)
}

func (db *Database) GetType() string {
	return "Database"
}

func (db *Database) String() string {
	return fmt.Sprintf("Database[ID=%s, Connections=%d, CostPerHr=%.2f, PreviousCostPerHr=%.2f, Owner=%s]", db.ID, db.Connections, db.CostPerHr, db.PreviousCostPerHr, db.Owner)
}
