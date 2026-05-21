import { proxyActivities } from '@temporalio/workflow';
import type { MyActivities } from './activities';

const { doWork } = proxyActivities<MyActivities>({
  startToCloseTimeout: '5m',
});

export async function goodWorkflow(input: WorkflowInput): Promise<string> {
  return await doWork(input);
}
