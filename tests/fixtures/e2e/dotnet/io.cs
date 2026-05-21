using Temporalio.Workflows;

[Workflow]
public class IoWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        var content = File.ReadAllText("/tmp/data.txt");
    }
}
