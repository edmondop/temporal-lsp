use temporal_sdk::workflow;
use std::thread;
use std::time::Duration;

#[workflow]
async fn delayed_workflow() -> Result<(), anyhow::Error> {
    thread::sleep(Duration::from_secs(5));
    Ok(())
}
