use temporal_sdk::workflow;

#[workflow]
async fn polling_workflow() -> Result<(), anyhow::Error> {
    loop {
        check_status().await;
    }
}
