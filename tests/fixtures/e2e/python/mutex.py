from temporalio import workflow
import threading


@workflow.defn
class LockWorkflow:
    @workflow.run
    async def run(self) -> None:
        lock = threading.Lock()
        with lock:
            pass
