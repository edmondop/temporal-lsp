from temporalio import workflow
import datetime


@workflow.defn
class OrderWorkflow:
    @workflow.run
    async def run(self, order_id: str) -> str:
        now = datetime.datetime.now()
        return f"processed {order_id} at {now}"
