package modules

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Module struct {
	Name string
	Path string
	Cmd  *exec.Cmd
}

type ModuleManager struct {
	Modules []Module
}

// Constructor
func NewModuleManager() *ModuleManager {
	return &ModuleManager{}
}

// resolveModulePath helps find the binary during development
func resolveModulePath(name string) string {
	// 1. Check relative path in Rust target directory (release)
	path, _ := filepath.Abs(filepath.Join("..", "modules", name, "target", "release", name+".exe"))
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// 2. Check relative path in Rust target directory (debug)
	path, _ = filepath.Abs(filepath.Join("..", "modules", name, "target", "debug", name+".exe"))
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// 3. Check current directory
	path, _ = filepath.Abs(name + ".exe")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return name // Fallback to PATH lookup
}

// Start modules
func (m *ModuleManager) StartModules(modules []Module) {
	m.Modules = modules
	for i := range m.Modules {
		mod := &m.Modules[i]
		log.Printf("Starting module: %s", mod.Name)

		binPath := resolveModulePath(mod.Name)
		cmd := exec.Command(binPath)

		// Set working directory to module root if possible, or agent root
		// For now, let's keep it in agent root or module dir?
		// keeping it default (Agent root) is usually safer for config loading etc.

		// Inherit stdout/stderr for logging (or capture it?)
		// For now, let's pipe it to supervisor stdout
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start module %s at %s: %v", mod.Name, binPath, err)
			continue
		}

		mod.Cmd = cmd
		log.Printf("Module %s started with PID %d", mod.Name, cmd.Process.Pid)
	}
}

// Stop modules
func (m *ModuleManager) StopModules() {
	for _, mod := range m.Modules {
		if mod.Cmd != nil && mod.Cmd.Process != nil {
			log.Printf("Stopping module: %s (PID %d)", mod.Name, mod.Cmd.Process.Pid)
			if err := mod.Cmd.Process.Kill(); err != nil {
				log.Printf("Failed to kill module %s: %v", mod.Name, err)
			}
		}
	}
	m.Modules = nil
}
