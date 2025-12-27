//go:build !windows && !linux && !darwin

package modules

import "os/exec"

func (mod *Module) applyPlatformResourceLimits(cmd *exec.Cmd) error {
	return nil
}

func (mod *Module) closePlatformResourceGovernor() {
}
