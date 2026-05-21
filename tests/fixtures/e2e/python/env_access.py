from temporalio import workflow
import os


@workflow.defn
class ConfigWorkflow:
    @workflow.run
    async def run(self) -> str:
        return os.getenv("API_KEY", "")
