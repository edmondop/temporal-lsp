import { proxyActivities } from '@temporalio/workflow';

export async function loggingWorkflow(): Promise<void> {
  console.log('processing...');
}
