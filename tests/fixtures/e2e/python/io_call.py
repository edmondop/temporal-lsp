from temporalio import workflow
import requests


@workflow.defn
class FetchWorkflow:
    @workflow.run
    async def run(self, url: str) -> str:
        response = requests.get(url)
        return response.text
