package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jsumanpan29/AegisAgent/internal/supervisor"
	"github.com/kardianos/service"
)

func main() {
	svcConfig := &service.Config{
		Name:        "AegisAgent",
		DisplayName: "Aegis Agent Supervisor",
		Description: "Supervisor service for AegisAgent monitoring and modules",
	}

	prog := &supervisor.Program{}

	s, err := service.New(prog, svcConfig)
	if err != nil {
		fmt.Println("Error creating service:", err)
		return
	}

	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			fmt.Println("Service control error:", err)
		} else {
			fmt.Println("Service command executed:", os.Args[1])
		}
		return
	}

	err = s.Run()
	if err != nil {
		log.Println("Service failed:", err)
	}
}
