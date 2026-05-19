from temporalio import workflow
import datetime
import time
import random
import requests
import threading
import queue


@workflow.defn
class BadWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        now = datetime.datetime.now()
        time.sleep(1)
        val = random.randint(1, 10)
        resp = requests.get("http://example.com")
        t = threading.Thread(target=lambda: None)
        lock = threading.Lock()
        q = queue.Queue()
        return str(val)
