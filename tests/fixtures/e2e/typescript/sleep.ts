import { proxyActivities } from '@temporalio/workflow';

export async function delayedWorkflow(): Promise<void> {
  setTimeout(() => {}, 5000);
}
