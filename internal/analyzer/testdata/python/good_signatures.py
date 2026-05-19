from temporalio import workflow, activity
from dataclasses import dataclass


@dataclass
class WorkflowInput:
    name: str
    age: int
    active: bool


@dataclass
class ActivityInput:
    name: str
    count: int


@workflow.defn
class GoodSignatureWorkflow:
    @workflow.run
    async def run(self, input: WorkflowInput) -> str:
        return f"{input.name}-{input.age}"


@activity.defn
async def good_activity(input: ActivityInput) -> str:
    return f"{input.name}-{input.count}"
