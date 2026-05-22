using Temporalio.Workflows;

[Workflow]
public class RandomWorkflow
{
    [WorkflowRun]
    public async Task<int> RunAsync()
    {
        var rng = new Random();
        return rng.Next(100);
    }
}
