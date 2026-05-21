use temporal_sdk::workflow;

#[workflow]
async fn greeting_workflow(name: String) -> Result<String, anyhow::Error> {
    let opts = ActivityOptions {
        start_to_close_timeout: Duration::from_secs(10),
        ..Default::default()
    };
    let result = execute_activity(greet_activity, name, opts).await?;
    Ok(result)
}
