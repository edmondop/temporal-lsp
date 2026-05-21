import { proxyActivities } from '@temporalio/workflow';

export async function envWorkflow(): Promise<string> {
  return process.env.HOME || '';
}
