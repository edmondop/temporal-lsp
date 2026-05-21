using Temporalio.Workflows;

[Workflow]
public class LoggingWorkflow
{
    [WorkflowRun]
    public async Task RunAsync()
    {
        Console.WriteLine("Processing...");
    }
}
