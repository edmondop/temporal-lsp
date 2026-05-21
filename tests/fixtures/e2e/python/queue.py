from temporalio import workflow
import queue


@workflow.defn
class PipelineWorkflow:
    @workflow.run
    async def run(self) -> None:
        q = queue.Queue()
        q.put("task")
