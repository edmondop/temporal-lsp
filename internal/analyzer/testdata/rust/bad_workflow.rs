use temporal_sdk::prelude::*;
use std::time::{SystemTime, Instant};
use std::thread;
use std::sync::Mutex;

#[workflow_run]
pub async fn bad_workflow(ctx: WfContext) -> Result<String, anyhow::Error> {
    let now = SystemTime::now();
    let instant = Instant::now();
    thread::sleep(std::time::Duration::from_secs(1));
    let rng = rand::thread_rng();
    let data = std::fs::read_to_string("file.txt")?;
    thread::spawn(|| {});
    let lock = Mutex::new(0);
    Ok("done".to_string())
}
