from temporalio import workflow
import os


@workflow.defn
class BadEnvWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        # Reading environment variables breaks determinism
        db_url = os.getenv("DATABASE_URL")
        secret = os.environ.get("API_KEY")
        home = os.environ["HOME"]
        return db_url or ""
