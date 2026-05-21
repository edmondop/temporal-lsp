import { proxyActivities } from '@temporalio/workflow';

export async function myWorkflow(): Promise<string> {
  const now = Date.now();
  return now.toString();
}
