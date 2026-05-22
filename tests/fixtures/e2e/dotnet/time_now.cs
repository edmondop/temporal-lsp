using Temporalio.Workflows;

[Workflow]
public class MyWorkflow
{
    [WorkflowRun]
    public async Task<string> RunAsync()
    {
        var now = DateTime.Now;
        return now.ToString();
    }
}
