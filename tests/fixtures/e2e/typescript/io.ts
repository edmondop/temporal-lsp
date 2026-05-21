import { proxyActivities } from '@temporalio/workflow';

export async function ioWorkflow(): Promise<void> {
  await fetch('https://example.com/api');
}
