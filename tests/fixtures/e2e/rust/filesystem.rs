use temporal_sdk::workflow;

#[workflow]
async fn backup_workflow(path: String) -> Result<String, anyhow::Error> {
    let contents = std::fs::read_to_string(&path)?;
    Ok(contents)
}
