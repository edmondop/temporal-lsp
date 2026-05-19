from temporalio import workflow
import datetime
import time


@workflow.defn
class BadTimeWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        # Non-deterministic: replays will get different times
        now = datetime.datetime.now()
        time.sleep(1)
        return str(now)
