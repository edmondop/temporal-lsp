using Temporalio.Workflows;

[Workflow]
public class MutexWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        var sem = new SemaphoreSlim(1);
        await sem.WaitAsync();
    }
}
