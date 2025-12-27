//go:build linux

package modules

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const cgroupRoot = "/sys/fs/cgroup/aegis-agent"

func (mod *Module) applyPlatformResourceLimits(cmd *exec.Cmd) error {
	cgPath := filepath.Join(cgroupRoot, mod.Name)

	// Ensure cgroup exists
	if err := os.MkdirAll(cgPath, 0755); err != nil {
		return fmt.Errorf("failed to create cgroup: %v", err)
	}

	// Set CPU limit (5% of 100000 period = 5000)
	if err := os.WriteFile(filepath.Join(cgPath, "cpu.max"), []byte("5000 100000"), 0644); err != nil {
		log.Printf("[Warning] Failed to set Linux CPU limit: %v", err)
	}

	// Set Memory limit (200MB)
	if err := os.WriteFile(filepath.Join(cgPath, "memory.max"), []byte(strconv.FormatInt(200*1024*1024, 10)), 0644); err != nil {
		log.Printf("[Warning] Failed to set Linux memory limit: %v", err)
	}

	// Assign PID to cgroup
	pidStr := strconv.Itoa(cmd.Process.Pid)
	return os.WriteFile(filepath.Join(cgPath, "cgroup.procs"), []byte(pidStr), 0644)
}

func (mod *Module) closePlatformResourceGovernor() {
	// Cgroup cleanup is often handled by systemd or manual removal if empty.
	// For now, we leave the cgroup directory.
}
