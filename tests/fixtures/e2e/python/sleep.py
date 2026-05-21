from temporalio import workflow
import time


@workflow.defn
class DelayWorkflow:
    @workflow.run
    async def run(self) -> None:
        time.sleep(5)
