use temporal_sdk::prelude::*;

#[workflow_run]
pub async fn config_workflow(ctx: WfContext) -> Result<String, anyhow::Error> {
    let config = std::fs::read_to_string("/etc/app/config.toml")?;
    Ok(config)
}
