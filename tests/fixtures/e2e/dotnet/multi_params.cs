using Temporalio.Workflows;

[Workflow]
public class MultiParamWorkflow
{
    [WorkflowRun]
    public async Task<string> RunAsync(string name, int age, bool active)
    {
        return $"{name}-{age}-{active}";
    }
}
