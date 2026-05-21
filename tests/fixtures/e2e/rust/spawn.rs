use temporal_sdk::workflow;

#[workflow]
async fn parallel_workflow() -> Result<(), anyhow::Error> {
    tokio::spawn(async { do_work().await });
    Ok(())
}
