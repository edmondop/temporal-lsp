from temporalio import workflow
from datetime import timedelta


@workflow.defn
class GoodWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        await workflow.sleep(timedelta(seconds=1))
        return "hello"
