using Temporalio.Workflows;

[Workflow]
public class ThreadWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        Task.Run(() => DoWork());
    }
}
