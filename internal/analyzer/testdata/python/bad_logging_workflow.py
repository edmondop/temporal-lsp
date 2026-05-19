from temporalio import workflow
import logging


logger = logging.getLogger(__name__)


@workflow.defn
class BadLoggingWorkflow:
    @workflow.run
    async def run(self, input: str) -> str:
        # Standard logging replays duplicate messages; use workflow.logger
        logging.info("starting workflow")
        logger.warning("processing %s", input)
        print("debug output")
        return "done"
