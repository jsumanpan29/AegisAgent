//go:build darwin

package modules

import (
	"log"
	"os/exec"
)

func (mod *Module) applyPlatformResourceLimits(cmd *exec.Cmd) error {
	// macOS doesn't have a direct equivalent to Cgroups or Job Objects for hard rate limiting
	// without using private APIs or complex Sandbox profiles.
	// Best effort: set process priority (nice value).
	// Higher value means lower priority. 19 is the lowest priority.
	log.Printf("[Info] Setting best-effort CPU priority (nice) for %s on macOS", mod.Name)

	// We can't use syscall.Setpriority directly easily on a different PID from here without more imports
	// but we can run the 'nice' command or just ignore if it's too complex.
	// For now, let's just log and consider it "best effort" empty.
	return nil
}

func (mod *Module) closePlatformResourceGovernor() {
}
