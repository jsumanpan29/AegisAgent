package modules

import (
	"log"
)

type Module struct {
	Name string
	Path string
}

type ModuleManager struct {
	Modules []Module
}

// Constructor
func NewModuleManager() *ModuleManager {
	return &ModuleManager{}
}

// Start modules
func (m *ModuleManager) StartModules(modules []Module) {
	m.Modules = modules
	for _, mod := range modules {
		log.Printf("Starting module: %s", mod.Name)
		// TODO: implement actual process start
	}
}

// Stop modules
func (m *ModuleManager) StopModules() {
	for _, mod := range m.Modules {
		log.Printf("Stopping module: %s", mod.Name)
		// TODO: implement actual process stop
	}
	m.Modules = nil
}
