//go:build windows

package modules

import (
	"log"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// Missing from x/sys/windows
	JOB_OBJECT_CPU_RATE_CONTROL_ENABLE   = 0x00000001
	JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP = 0x00000004
	JobObjectCpuRateControlInformation   = 15
)

type JOBOBJECT_CPU_RATE_CONTROL_INFORMATION struct {
	ControlFlags uint32
	CpuRate      uint32
}

func (mod *Module) applyPlatformResourceLimits(cmd *exec.Cmd) error {
	if mod.Extra.JobHandle == 0 {
		h, err := windows.CreateJobObject(nil, nil)
		if err != nil {
			return err
		}
		mod.Extra.JobHandle = uintptr(h)

		// Set limits: 5% CPU
		cpuLimit := JOBOBJECT_CPU_RATE_CONTROL_INFORMATION{
			ControlFlags: JOB_OBJECT_CPU_RATE_CONTROL_ENABLE | JOB_OBJECT_CPU_RATE_CONTROL_HARD_CAP,
			CpuRate:      500, // 5%
		}

		_, err = windows.SetInformationJobObject(
			windows.Handle(mod.Extra.JobHandle),
			uint32(JobObjectCpuRateControlInformation),
			uintptr(unsafe.Pointer(&cpuLimit)),
			uint32(unsafe.Sizeof(cpuLimit)),
		)
		if err != nil {
			log.Printf("[Warning] Failed to set CPU limit for %s: %v", mod.Name, err)
		}

		// Set limits: 200MB RAM
		memLimit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
			BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
				LimitFlags: windows.JOB_OBJECT_LIMIT_JOB_MEMORY | windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
			},
			JobMemoryLimit: 200 * 1024 * 1024,
		}

		_, err = windows.SetInformationJobObject(
			windows.Handle(mod.Extra.JobHandle),
			windows.JobObjectExtendedLimitInformation,
			uintptr(unsafe.Pointer(&memLimit)),
			uint32(unsafe.Sizeof(memLimit)),
		)
		if err != nil {
			log.Printf("[Warning] Failed to set Memory limit for %s: %v", mod.Name, err)
		}
	}

	hProcess, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(hProcess)

	return windows.AssignProcessToJobObject(windows.Handle(mod.Extra.JobHandle), hProcess)
}

func (mod *Module) closePlatformResourceGovernor() {
	if mod.Extra.JobHandle != 0 {
		windows.CloseHandle(windows.Handle(mod.Extra.JobHandle))
		mod.Extra.JobHandle = 0
	}
}
