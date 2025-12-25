package supervisor

import (
	"log"
	"time"

	"github.com/jsumanpan29/AegisAgent/internal/config"
	"github.com/jsumanpan29/AegisAgent/internal/ipc"
	"github.com/jsumanpan29/AegisAgent/internal/modules"

	"github.com/kardianos/service"
)

type Program struct {
	exit         chan struct{}
	moduleMgr    *modules.ModuleManager
	config       *config.Config
	ipcInterface ipc.IPC
}

func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	log.Println("Aegis Agent service starting...")

	// Load modules manager
	p.moduleMgr = &modules.ModuleManager{}
	if p.config != nil {
		p.moduleMgr.StartModules(p.config.Modules)
	}

	// TODO: Initialize IPC
	// p.ipcInterface = ipc.NewNamedPipeIPC("AegisPipe")

	go p.run()
	return nil
}

func (p *Program) run() {
	log.Println("Aegis Agent main loop running...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Aegis Agent heartbeat...")
			// TODO: Poll IPC, forward events, check modules
		case <-p.exit:
			log.Println("Aegis Agent shutting down main loop")
			return
		}
	}
}

func (p *Program) Stop(s service.Service) error {
	close(p.exit)
	log.Println("Aegis Agent stopping...")

	if p.moduleMgr != nil {
		p.moduleMgr.StopModules()
	}

	if p.ipcInterface != nil {
		p.ipcInterface.Close()
	}

	return nil
}
