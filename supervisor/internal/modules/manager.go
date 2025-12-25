package modules

import (
	"log"
)

type ModuleManager struct{}

func (m *ModuleManager) StartModules(modules []string) {
	log.Println("Starting modules:", modules)
	// TODO: implement module start logic
}

func (m *ModuleManager) StopModules() {
	log.Println("Stopping all modules")
	// TODO: implement module stop logic
}
