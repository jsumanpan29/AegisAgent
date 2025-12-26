use log::{info, error, warn};
use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::Duration;
use sysinfo::{System, Pid, ProcessRefreshKind};

#[cfg(windows)]
use std::ffi::OsStr;
#[cfg(windows)]
use std::os::windows::ffi::OsStrExt;
#[cfg(windows)]
use winapi::um::fileapi::{CreateFileW, OPEN_EXISTING};
#[cfg(windows)]
use winapi::um::handleapi::{CloseHandle, INVALID_HANDLE_VALUE};
#[cfg(windows)]
use winapi::um::winbase::FILE_FLAG_OVERLAPPED;
#[cfg(windows)]
use winapi::um::winnt::{GENERIC_READ, GENERIC_WRITE, FILE_SHARE_READ, FILE_SHARE_WRITE, HANDLE};

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ProcessInfo {
    pid: u32,
    name: String,
    cpu_usage: f32,
    memory_kb: u64,
    timestamp: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct MonitorReport {
    module: String,
    process_count: usize,
    top_processes: Vec<ProcessInfo>,
    timestamp: String,
}

#[cfg(windows)]
struct NamedPipeClient {
    handle: HANDLE,
}

#[cfg(windows)]
impl NamedPipeClient {
    fn connect(pipe_name: &str) -> Result<Self, String> {
        let wide_name: Vec<u16> = OsStr::new(pipe_name)
            .encode_wide()
            .chain(std::iter::once(0))
            .collect();

        unsafe {
            let handle = CreateFileW(
                wide_name.as_ptr(),
                GENERIC_READ | GENERIC_WRITE,
                FILE_SHARE_READ | FILE_SHARE_WRITE,
                std::ptr::null_mut(),
                OPEN_EXISTING,
                FILE_FLAG_OVERLAPPED,
                std::ptr::null_mut(),
            );

            if handle == INVALID_HANDLE_VALUE {
                return Err(format!("Failed to connect to pipe: {}", pipe_name));
            }

            Ok(NamedPipeClient { handle })
        }
    }

    fn send_message(&self, message: &str) -> Result<(), String> {
        use winapi::um::fileapi::WriteFile;
        
        let bytes = message.as_bytes();
        let mut bytes_written: u32 = 0;

        unsafe {
            let result = WriteFile(
                self.handle,
                bytes.as_ptr() as *const _,
                bytes.len() as u32,
                &mut bytes_written,
                std::ptr::null_mut(),
            );

            if result == 0 {
                return Err("Failed to write to pipe".to_string());
            }
        }

        Ok(())
    }
}

#[cfg(windows)]
impl Drop for NamedPipeClient {
    fn drop(&mut self) {
        unsafe {
            CloseHandle(self.handle);
        }
    }
}

fn collect_process_info(sys: &mut System) -> Vec<ProcessInfo> {
    sys.refresh_processes_specifics(ProcessRefreshKind::everything());
    
    let mut processes: Vec<ProcessInfo> = sys
        .processes()
        .iter()
        .map(|(pid, process)| ProcessInfo {
            pid: pid.as_u32(),
            name: process.name().to_string(),
            cpu_usage: process.cpu_usage(),
            memory_kb: process.memory() / 1024,
            timestamp: chrono::Local::now().to_rfc3339(),
        })
        .collect();

    // Sort by CPU usage and take top 10
    processes.sort_by(|a, b| b.cpu_usage.partial_cmp(&a.cpu_usage).unwrap());
    processes.truncate(10);
    
    processes
}

fn main() {
    env_logger::Builder::from_default_env()
        .filter_level(log::LevelFilter::Info)
        .init();

    info!("Process Monitor Module starting...");

    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();

    // Handle Ctrl+C for graceful shutdown
    ctrlc::set_handler(move || {
        warn!("Received shutdown signal");
        r.store(false, Ordering::SeqCst);
    })
    .expect("Error setting Ctrl-C handler");

    let mut sys = System::new_all();
    let pipe_name = r"\\.\pipe\AegisPipe_process_monitor";

    #[cfg(windows)]
    let pipe_client = match NamedPipeClient::connect(pipe_name) {
        Ok(client) => {
            info!("Connected to supervisor via Named Pipe: {}", pipe_name);
            Some(client)
        }
        Err(e) => {
            warn!("Could not connect to supervisor pipe: {}. Running in standalone mode.", e);
            None
        }
    };

    #[cfg(not(windows))]
    let pipe_client: Option<()> = None;

    let mut iteration = 0;
    
    while running.load(Ordering::SeqCst) {
        iteration += 1;
        
        let top_processes = collect_process_info(&mut sys);
        let total_processes = sys.processes().len();

        let report = MonitorReport {
            module: "process_monitor".to_string(),
            process_count: total_processes,
            top_processes: top_processes.clone(),
            timestamp: chrono::Local::now().to_rfc3339(),
        };

        info!(
            "Iteration {}: Monitoring {} processes, Top CPU: {} ({:.2}%)",
            iteration,
            total_processes,
            top_processes.first().map(|p| p.name.as_str()).unwrap_or("N/A"),
            top_processes.first().map(|p| p.cpu_usage).unwrap_or(0.0)
        );

        // Send report to supervisor if connected
        #[cfg(windows)]
        if let Some(ref client) = pipe_client {
            match serde_json::to_string(&report) {
                Ok(json) => {
                    if let Err(e) = client.send_message(&json) {
                        error!("Failed to send report to supervisor: {}", e);
                    } else {
                        info!("Report sent to supervisor");
                    }
                }
                Err(e) => error!("Failed to serialize report: {}", e),
            }
        }

        // Wait 5 seconds before next iteration
        thread::sleep(Duration::from_secs(5));
    }

    info!("Process Monitor Module shutting down gracefully");
}
