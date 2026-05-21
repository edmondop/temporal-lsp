using Temporalio.Workflows;

[Workflow]
public class PollingWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        while (true)
        {
            await CheckStatus();
        }
    }
}
