use temporal_sdk::prelude::*;

#[workflow_run]
pub async fn bad_workflow_sig(ctx: WfContext, name: String, age: u32, active: bool) -> Result<String, anyhow::Error> {
    Ok(format!("{} {} {}", name, age, active))
}

#[activity]
pub async fn bad_activity(ctx: ActContext, id: String, count: i32) -> Result<String, anyhow::Error> {
    Ok(format!("{} {}", id, count))
}
