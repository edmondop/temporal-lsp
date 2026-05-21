import { proxyActivities } from '@temporalio/workflow';

export async function pollingWorkflow(): Promise<void> {
  while (true) {
    await checkStatus();
  }
}
