package models

import (
	"fmt"
	"math/rand"
)

func (vm *VM) UpdateUsage() {
	vm.CPUUsage = rand.Float64() * 100
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
	return fmt.Sprintf("VM[ID=%s, CPUUsage=%.2f, CostPerHour=%.2f]", vm.ID, vm.CPUUsage, vm.CostPerHour)
}
