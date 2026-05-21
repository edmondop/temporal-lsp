use temporal_sdk::workflow;

#[workflow]
async fn lottery_workflow() -> Result<u32, anyhow::Error> {
    let rng = rand::thread_rng();
    Ok(rng.gen_range(1..100))
}
