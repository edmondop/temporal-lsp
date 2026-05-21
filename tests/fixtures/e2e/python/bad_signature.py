from temporalio import workflow, activity


@workflow.defn
class MultiParamWorkflow:
    @workflow.run
    async def run(self, name: str, age: int, active: bool) -> str:
        return f"{name} is {age}"


@activity.defn
async def send_email(to: str, subject: str, body: str) -> bool:
    return True
