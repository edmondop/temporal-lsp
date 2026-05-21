from temporalio import workflow
import random


@workflow.defn
class LotteryWorkflow:
    @workflow.run
    async def run(self) -> int:
        return random.randint(1, 100)
