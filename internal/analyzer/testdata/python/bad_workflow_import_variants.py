from temporalio.workflow import defn, run
import datetime
import time


@defn
class BadWorkflowDirectImport:
    @run
    async def run(self, input: str) -> str:
        now = datetime.datetime.now()
        time.sleep(1)
        return str(now)
