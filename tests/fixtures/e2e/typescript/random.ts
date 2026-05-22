import { proxyActivities } from '@temporalio/workflow';

export async function randomWorkflow(): Promise<number> {
  return Math.random();
}
