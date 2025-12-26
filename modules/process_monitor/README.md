# Process Monitor Module

A Rust-based process monitoring module for AegisAgent that collects system process information and communicates with the Go supervisor via Named Pipes (Windows).

## Features

- **Process Monitoring**: Tracks all running processes with PID, name, CPU usage, and memory consumption
- **Top Process Reporting**: Reports the top 10 processes by CPU usage
- **IPC Communication**: Communicates with the Go supervisor via Named Pipes (`\\.\pipe\AegisPipe_process_monitor`)
- **Graceful Shutdown**: Handles SIGINT/SIGTERM for clean termination
- **Standalone Mode**: Can run independently if supervisor is not available

## Building

```bash
cd modules/process_monitor
cargo build --release
```

The compiled binary will be located at `target/release/process_monitor.exe`

## Running

### Standalone Mode
```bash
cargo run
```

### With Supervisor
The Go supervisor will automatically launch this module when configured in `config/agent.yml`:
```yaml
modules:
  - process_monitor
```

## IPC Protocol

The module sends JSON reports to the supervisor every 5 seconds:

```json
{
  "module": "process_monitor",
  "process_count": 150,
  "top_processes": [
    {
      "pid": 1234,
      "name": "chrome.exe",
      "cpu_usage": 15.5,
      "memory_kb": 524288,
      "timestamp": "2025-12-26T17:00:00+08:00"
    }
  ],
  "timestamp": "2025-12-26T17:00:00+08:00"
}
```

## Dependencies

- `sysinfo`: Cross-platform system information library
- `serde` / `serde_json`: JSON serialization
- `winapi`: Windows API bindings for Named Pipes
- `log` / `env_logger`: Logging infrastructure
- `chrono`: Timestamp generation

## Configuration

Set the `RUST_LOG` environment variable to control logging verbosity:
```bash
set RUST_LOG=info
cargo run
```

Levels: `error`, `warn`, `info`, `debug`, `trace`
