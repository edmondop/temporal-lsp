using Temporalio.Workflows;

[Workflow]
public class EnvWorkflow
{
    [WorkflowRun]
    public async Task<string> RunAsync()
    {
        return Environment.GetEnvironmentVariable("HOME");
    }
}
