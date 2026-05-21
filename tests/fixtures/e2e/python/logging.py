from temporalio import workflow
import logging

logger = logging.getLogger(__name__)


@workflow.defn
class VerboseWorkflow:
    @workflow.run
    async def run(self) -> None:
        logging.info("starting workflow")
        logger.debug("debug info")
        print("hello")
