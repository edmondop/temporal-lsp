import { proxyActivities } from '@temporalio/workflow';

export async function multiParamWorkflow(name: string, age: number, active: boolean): Promise<string> {
  return `${name}-${age}-${active}`;
}
