using Temporalio.Workflows;

[Workflow]
public class SleepWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        Thread.Sleep(5000);
        await Task.Delay(TimeSpan.FromSeconds(1));
    }
}
