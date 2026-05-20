use temporal_sdk::prelude::*;

#[workflow_run]
pub async fn good_workflow(ctx: WfContext) -> Result<String, anyhow::Error> {
    let result = ctx.activity(MyActivityOptions::default()).await?;
    ctx.timer(std::time::Duration::from_secs(1)).await;
    Ok(result)
}
