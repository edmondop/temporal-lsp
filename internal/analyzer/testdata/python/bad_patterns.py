from temporalio import workflow, activity
from datetime import timedelta


@workflow.defn
class BadPatternWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        # activity-timeout-required: no timeout passed
        result = await workflow.execute_activity(
            "my_activity",
            input,
            schedule_to_close_timeout=None,
        )

        # unbounded-loop: while True without continue_as_new
        while True:
            await workflow.sleep(timedelta(seconds=1))

        return result


@activity.defn
async def bad_activity(input: str) -> str:
    return input
