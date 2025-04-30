package models

import (
	"fmt"
	"math/rand"
	"time"
)

func (vm *VM) UpdateUsage() {
	vm.CPUUsage = rand.Float64() * 100
	if rand.Float64() < 0.2 {
		vm.LastActive = time.Now().Unix()
	}
}

func (vm *VM) GetId() string {
	return vm.ID
}

func (vm *VM) GetUsage() float64 {
	return vm.CPUUsage
}

func (vm *VM) GetType() string {
	return "VM"
}

func (vm *VM) String() string {
	return fmt.Sprintf("VM[ID=%s, CPUUsage=%.2f, CostPerHour=%.2f, PreviousCostPerHour=%.2f, Owner=%s, LastActive=%d]", vm.ID, vm.CPUUsage, vm.CostPerHour, vm.PreviousCostPerHour, vm.Owner, vm.LastActive)
}
