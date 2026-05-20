use temporal_sdk::prelude::*;

#[workflow_run]
pub async fn bad_patterns_workflow(ctx: WfContext) -> Result<String, anyhow::Error> {
    ctx.execute_activity(my_activity).await?;
    loop {
        ctx.timer(std::time::Duration::from_secs(60)).await;
    }
}
