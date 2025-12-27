package modules

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Module struct {
	Name            string
	Path            string
	Cmd             *exec.Cmd
	RestartCount    int
	LastRestart     time.Time
	TrackedCrashes  []time.Time
	stopChan        chan struct{}
	LastPingSuccess time.Time
	Extra           PlatformSpecificData
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
		mod.stopChan = make(chan struct{})
		mod.TrackedCrashes = make([]time.Time, 0)
		// Start supervision loop in background
		go m.supervise(mod)
	}
}

func (mod *Module) isCrashLooping() bool {
	now := time.Now()
	threshold := 10 * time.Second
	maxCrashes := 5

	// Filter crashes within the last 10 seconds
	var recent []time.Time
	for _, t := range mod.TrackedCrashes {
		if now.Sub(t) < threshold {
			recent = append(recent, t)
		}
	}
	mod.TrackedCrashes = recent

	return len(mod.TrackedCrashes) >= maxCrashes
}

func (m *ModuleManager) supervise(mod *Module) {
	backoff := time.Second

	for {
		select {
		case <-mod.stopChan:
			return
		default:
			// proceed
		}

		binPath := resolveModulePath(mod.Name)
		cmd := exec.Command(binPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Use SysProcAttr to start process suspended, so we can add it to Job Object before it runs?
		// Actually, AssignProcessToJobObject works on running processes too.
		// But for strict limits, sometimes it's better to start suspended.
		// For now, let's keep it simple.

		mod.Cmd = cmd

		log.Printf("Starting module: %s", mod.Name)
		mod.LastRestart = time.Now()

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start module %s: %v", mod.Name, err)
		} else {
			log.Printf("Module %s started (PID %d)", mod.Name, cmd.Process.Pid)

			// Assign to Resource Governor (Job Object on Windows, Cgroups on Linux)
			if err := mod.ApplyResourceLimits(cmd); err != nil {
				log.Printf("[Warning] Failed to apply resource limits for %s: %v", mod.Name, err)
			} else {
				log.Printf("Module %s assigned to Resource Governor", mod.Name)
			}

			// Monitor process until exit
			err := cmd.Wait()
			log.Printf("Module %s exited: %v", mod.Name, err)
		}

		mod.TrackedCrashes = append(mod.TrackedCrashes, time.Now())

		if mod.isCrashLooping() {
			log.Printf("[CRITICAL] Module %s is in a crash loop! Disabling.", mod.Name)
			return
		}

		// Check if we should stop (in case stop signal came while waiting)
		select {
		case <-mod.stopChan:
			return
		default:
			// continue to backoff
		}

		// Reset backoff if module ran for more than 5 minutes
		if time.Since(mod.LastRestart) > 5*time.Minute {
			backoff = time.Second
			mod.RestartCount = 0
		}

		mod.RestartCount++
		log.Printf("[Watchdog] Restarting module %s in %v (Attempt %d)", mod.Name, backoff, mod.RestartCount)

		select {
		case <-mod.stopChan:
			return
		case <-time.After(backoff):
			// continue loop
		}

		// Exponential backoff with 60s cap
		backoff *= 2
		if backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}
}

// Stop modules
func (m *ModuleManager) StopModules() {
	for i := range m.Modules {
		mod := &m.Modules[i]

		// Signal loop to stop
		if mod.stopChan != nil {
			close(mod.stopChan)
		}

		// Kill process immediately
		if mod.Cmd != nil && mod.Cmd.Process != nil {
			log.Printf("Stopping module: %s (PID %d)", mod.Name, mod.Cmd.Process.Pid)
			if err := mod.Cmd.Process.Kill(); err != nil {
				log.Printf("Failed to kill module %s: %v", mod.Name, err)
			}
		}

		mod.CloseResourceGovernor()
	}
	m.Modules = nil
}
