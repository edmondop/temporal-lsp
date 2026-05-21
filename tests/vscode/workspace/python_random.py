from temporalio import workflow
import random


@workflow.defn
class LotteryWorkflow:
    @workflow.run
    async def run(self, player_id: str) -> int:
        winning_number = random.randint(1, 100)
        return winning_number
