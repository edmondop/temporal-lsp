use temporal_sdk::workflow;
use std::sync::Mutex;

#[workflow]
async fn sync_workflow() -> Result<(), anyhow::Error> {
    let counter = Mutex::new(0);
    *counter.lock().unwrap() += 1;
    Ok(())
}
