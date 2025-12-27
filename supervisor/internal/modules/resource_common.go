package modules

import (
	"os/exec"
)

// PlatformSpecificData holds handles or identifiers used by the resource governor.
type PlatformSpecificData struct {
	JobHandle   uintptr
	linuxCgroup string
}

// ApplyResourceLimits sets up the platform-specific resource governor (Job Objects on Windows, Cgroups on Linux).
// It returns an error if the governor could not be initialized or applied to the process.
func (mod *Module) ApplyResourceLimits(cmd *exec.Cmd) error {
	return mod.applyPlatformResourceLimits(cmd)
}

// CloseResourceGovernor cleans up any resources associated with the governor.
func (mod *Module) CloseResourceGovernor() {
	mod.closePlatformResourceGovernor()
}
