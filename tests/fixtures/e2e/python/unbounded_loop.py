from temporalio import workflow
from datetime import timedelta


@workflow.defn
class PollingWorkflow:
    @workflow.run
    async def run(self) -> None:
        while True:
            await workflow.execute_activity(
                poll_status,
                start_to_close_timeout=timedelta(seconds=5),
            )
            await workflow.sleep(60)
