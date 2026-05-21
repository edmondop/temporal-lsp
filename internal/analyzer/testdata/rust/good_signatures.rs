use temporal_sdk::prelude::*;

#[derive(Serialize, Deserialize)]
pub struct WorkflowInput {
    pub name: String,
    pub age: u32,
}

#[workflow_run]
pub async fn good_workflow_sig(ctx: WfContext, input: WorkflowInput) -> Result<String, anyhow::Error> {
    Ok(format!("{} {}", input.name, input.age))
}

#[activity]
pub async fn good_activity(ctx: ActContext, input: ActivityInput) -> Result<String, anyhow::Error> {
    Ok(input.id)
}
