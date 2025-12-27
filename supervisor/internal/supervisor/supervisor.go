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

	// Initialize IPC clients per module
	p.ipcClients = make(map[string]ipc.IPC)
	if p.config != nil {
		for _, modName := range p.config.Modules {
			ipcName := "AegisPipe_" + modName
			client, err := ipc.NewIPC(ipcName)
			if err != nil {
				log.Printf("Failed to create IPC server for module %s: %v", modName, err)
				continue
			}
			p.ipcClients[modName] = client
			// Start goroutine to receive messages from this module
			go p.handleModuleMessages(modName, client)
		}
	}

	// Load modules manager
	p.moduleMgr = modules.NewModuleManager()
	if p.config != nil {
		// Convert []string to []modules.Module
		moduleList := make([]modules.Module, len(p.config.Modules))
		for i, modName := range p.config.Modules {
			moduleList[i] = modules.Module{
				Name:            modName,
				Path:            modName, // TODO: configure actual module paths
				LastPingSuccess: time.Now(),
			}
		}
		p.moduleMgr.StartModules(moduleList)
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
				text := string(msg)
				if text == "Pong" {
					// Update LastPingSuccess for this module
					for i := range p.moduleMgr.Modules {
						if p.moduleMgr.Modules[i].Name == moduleName {
							p.moduleMgr.Modules[i].LastPingSuccess = time.Now()
							break
						}
					}
				} else {
					log.Printf("[%s] Received message: %s", moduleName, text)
				}
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

			// Send Pings and check for timeouts
			for _, mod := range p.moduleMgr.Modules {
				if client, ok := p.ipcClients[mod.Name]; ok {
					// Send Ping
					_ = client.Send([]byte("Ping"))
				}

				// Check timeout (30s)
				if time.Since(mod.LastPingSuccess) > 30*time.Second {
					log.Printf("[CRITICAL] Liveness check failed for %s (No Pong for 30s). Restarting.", mod.Name)
					// We can't easily restart just one module from here without modifying manager.go
					// but for MVP, let's log it.
					// Actually, supervisor loop naturally restarts if process dies.
					// If it's deadlocked, we MUST kill it.
					if mod.Cmd != nil && mod.Cmd.Process != nil {
						mod.Cmd.Process.Kill()
					}
				}
			}
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
