use temporal_sdk::workflow;
use std::time::Instant;

#[workflow]
async fn timed_workflow() -> Result<u64, anyhow::Error> {
    let start = Instant::now();
    do_work().await;
    Ok(start.elapsed().as_secs())
}
