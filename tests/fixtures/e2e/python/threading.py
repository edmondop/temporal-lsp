from temporalio import workflow
import threading


@workflow.defn
class ParallelWorkflow:
    @workflow.run
    async def run(self) -> None:
        t = threading.Thread(target=self.do_work)
        t.start()
