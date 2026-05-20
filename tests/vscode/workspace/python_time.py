from temporalio import workflow
import datetime


@workflow.defn
class OrderWorkflow:
    @workflow.run
    async def run(self, order_id: str) -> str:
        created_at = datetime.datetime.now()
        return f"Order {order_id} created at {created_at}"
