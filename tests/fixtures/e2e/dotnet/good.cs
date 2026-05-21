using Temporalio.Workflows;

[Workflow]
public class GoodWorkflow
{
    [WorkflowRun]
    public async Task<string> RunAsync(MyInput input)
    {
        var result = await Workflow.ExecuteActivityAsync(
            () => MyActivities.DoWork(input),
            new() { StartToCloseTimeout = TimeSpan.FromMinutes(5) });
        return result;
    }
}
