# AegisAgent

**AegisAgent** is a modular system agent designed to manage and monitor Rust modules and kernel drivers via a Go-based supervisor. It supports Windows (MVP) and is structured to be cross-platform in the future. This project demonstrates Go service management, IPC communication, and modular architecture.

---

## Table of Contents

1. [Project Overview](#project-overview)  
2. [Architecture](#architecture)  
3. [Features](#features)  
4. [Folder Structure](#folder-structure)  
5. [Installation](#installation)  
6. [Running the Agent](#running-the-agent)  
7. [Configuration](#configuration)  
8. [Development](#development)  
9. [License](#license)

---

## Project Overview

AegisAgent is designed as a **supervisor service** in Go that:

- Orchestrates **Rust modules** and C++ kernel drivers  
- Communicates with modules via **IPC (Named Pipes on Windows, Unix Sockets on Linux/macOS)**  
- Runs as a **background service** with automatic start/stop and logging  

This project is ideal for demonstrating system programming, service orchestration, and cross-language integration in Go.

---

## Architecture
```
[Kernel Driver (C++)]
│
▼
IPC (Named Pipe / Unix Socket)
│
▼
[Go Supervisor / Aegis Service]
├─ Config Loader
├─ IPC Interface
├─ Module Manager (starts/stops Rust modules)
└─ Logging / Heartbeat
│
▼
[Rust Modules (processes)]
```
- **Supervisor**: main service written in Go, orchestrates modules  
- **Modules**: independent Rust binaries communicating via IPC  
- **Kernel driver**: platform-specific driver communicating with supervisor  
- **IPC**: abstracts communication across OS platforms  

---

## Features

- Go-based **Windows service**  
- Configurable via `agent.yml`  
- Logs heartbeat and module status  
- Modular, decoupled architecture for easy extension  
- IPC ready for cross-platform module communication  
- Supports hot-loading and monitoring Rust modules  

---

## Folder Structure
```
AegisAgent/
├─ cmd/supervisor/ # Main entry point for Go supervisor
│ └─ main.go
├─ internal/ # Go internal packages
│ ├─ config/
│ ├─ ipc/
│ ├─ logging/
│ ├─ modules/
│ └─ supervisor/
├─ modules/ # Rust module sources
│ └─ process_monitor/
├─ kernel/ # C++ kernel driver sources
├─ config/ # Default configuration files
│ └─ agent.yml
├─ docs/ # Architecture diagrams and notes
└─ .gitignore
```
---

## Installation

1. Clone the repository:

```bash
git clone https://github.com/jsumanpan29/AegisAgent.git
cd AegisAgent/supervisor
```
2. Build the executable:
```bash
go build -o aegis.exe ./cmd/supervisor
```

3. Install as Windows service (Admin privileges required):
```sh
aegis.exe install
aegis.exe start
```

## Running the Agent
- To run manually (for testing):
```sh
./aegis.exe
```

- To stop the service:
```sh
aegis.exe stop
```

- To uninstall service:
```sh
aegis.exe uninstall
```

- Logs are written to the path specified in agent.yml.

## Configuration:
```yml
modules:
  - process_monitor
log_path: "logs/agent.log"
heartbeat_interval: 5
```
- modules: list of modules to start
- log_path: log file location
- heartbeat_interval: time (seconds) between heartbeat logs

## Development
- Internal Go packages (internal/config, internal/ipc, internal/modules) handle modular responsibilities.
- Rust modules live in modules/ and communicate via IPC.
- Kernel driver resides in kernel/ folder.
- Use go build ./cmd/supervisor to compile the supervisor.
- Run tests by creating temporary modules or mocking IPC interfaces.

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
