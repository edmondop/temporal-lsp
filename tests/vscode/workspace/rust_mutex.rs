use temporal_sdk::prelude::*;
use std::sync::Mutex;

#[workflow_run]
pub async fn counter_workflow(ctx: WfContext) -> Result<u64, anyhow::Error> {
    let counter = Mutex::new(0u64);
    Ok(*counter.lock().unwrap())
}
