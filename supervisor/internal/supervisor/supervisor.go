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
	exit       chan struct{}
	moduleMgr  *modules.ModuleManager
	config     *config.Config
	ipcClients map[string]ipc.IPC // one IPC per module
}

func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	log.Println("Aegis Agent service starting...")

	// Load config
	cfg, err := config.LoadConfig("../config/agent.yml")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	} else {
		p.config = cfg
	}

	// Load modules manager
	p.moduleMgr = modules.NewModuleManager()
	if p.config != nil {
		// Convert []string to []modules.Module
		moduleList := make([]modules.Module, len(p.config.Modules))
		for i, modName := range p.config.Modules {
			moduleList[i] = modules.Module{
				Name: modName,
				Path: modName, // TODO: configure actual module paths
			}
		}
		p.moduleMgr.StartModules(moduleList)
	}

	// Initialize IPC clients per module
	p.ipcClients = make(map[string]ipc.IPC)
	if p.config != nil {
		for _, modName := range p.config.Modules {
			ipcName := "AegisPipe_" + modName
			client, err := ipc.NewIPC(ipcName)
			if err != nil {
				log.Printf("Failed to connect IPC for module %s: %v", modName, err)
				continue
			}
			p.ipcClients[modName] = client
			// Start goroutine to receive messages from this module
			go p.handleModuleMessages(modName, client)
		}
	}

	go p.run()
	return nil

}

func (p *Program) handleModuleMessages(moduleName string, client ipc.IPC) {
	for {
		select {
		case <-p.exit:
			return
		default:
			msg, err := client.Receive()
			if err != nil {
				// Handle or log error (pipe might be temporarily unavailable)
				continue
			}
			if msg != nil {
				log.Printf("[%s] Received message: %s", moduleName, string(msg))
			}
		}
	}
}

func (p *Program) run() {
	log.Println("Aegis Agent main loop running...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Aegis Agent heartbeat...")
			// TODO: Monitor module statuses, restart if necessary
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

	for name, client := range p.ipcClients {
		if client != nil {
			client.Close()
			log.Printf("Closed IPC for module %s", name)
		}
	}

	return nil

}
