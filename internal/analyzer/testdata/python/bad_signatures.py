from temporalio import workflow, activity


@workflow.defn
class BadSignatureWorkflow:
    @workflow.run
    async def run(self, name: str, age: int, active: bool) -> str:
        return f"{name}-{age}"


@activity.defn
async def bad_activity(name: str, age: int, count: int) -> str:
    return f"{name}-{age}-{count}"
