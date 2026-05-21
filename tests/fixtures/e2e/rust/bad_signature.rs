use temporal_sdk::workflow;

#[workflow]
async fn process_order(order_id: String, quantity: u32, price: f64) -> Result<String, anyhow::Error> {
    Ok(format!("{}: {} x {}", order_id, quantity, price))
}
