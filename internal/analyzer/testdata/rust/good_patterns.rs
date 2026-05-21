use temporal_sdk::prelude::*;
use std::time::Duration;

#[workflow_run]
pub async fn good_patterns_workflow(ctx: WfContext) -> Result<String, anyhow::Error> {
    let opts = ActivityOptions {
        start_to_close_timeout: Some(Duration::from_secs(30)),
        ..Default::default()
    };
    ctx.execute_activity(my_activity, opts).await?;
    loop {
        ctx.timer(Duration::from_secs(60)).await;
        ctx.continue_as_new()?;
    }
}
