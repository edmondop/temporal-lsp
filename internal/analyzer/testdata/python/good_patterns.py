from temporalio import workflow, activity
from datetime import timedelta


@workflow.defn
class GoodPatternWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        result = await workflow.execute_activity(
            "my_activity",
            input,
            start_to_close_timeout=timedelta(seconds=30),
        )

        # Bounded loop with continue_as_new
        count = 0
        while True:
            await workflow.sleep(timedelta(seconds=1))
            count += 1
            if count > 100:
                await workflow.continue_as_new(input)

        return result


@activity.defn
async def good_activity(input: str) -> str:
    return input
