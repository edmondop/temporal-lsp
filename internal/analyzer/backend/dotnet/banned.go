package dotnet

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

type bannedCall struct {
	Rule    rules.Rule
	Matches []string
}

var bannedCalls = []bannedCall{
	{
		Rule:    rules.TimeNow.WithMessage("Use Workflow.UtcNow instead of DateTime.Now/UtcNow in workflows"),
		Matches: []string{"DateTime.Now", "DateTime.UtcNow", "DateTimeOffset.Now", "DateTimeOffset.UtcNow"},
	},
	{
		Rule:    rules.Sleep.WithMessage("Use Workflow.DelayAsync() instead of Thread.Sleep or Task.Delay in workflows"),
		Matches: []string{"Thread.Sleep", "Task.Delay"},
	},
	{
		Rule:    rules.Random.WithMessage("Use Workflow.Random instead of System.Random in workflows"),
		Matches: []string{"new Random", "Random.Shared"},
	},
	{
		Rule:    rules.IO.WithMessage("Move IO operations to an activity"),
		Matches: []string{"File.", "Directory.", "StreamReader", "StreamWriter", "HttpClient", "WebClient"},
	},
	{
		Rule:    rules.Goroutine.WithMessage("Use Workflow.RunTaskAsync() instead of Task.Run or new Thread in workflows"),
		Matches: []string{"Task.Run", "new Thread", "ThreadPool.QueueUserWorkItem"},
	},
	{
		Rule:    rules.Mutex.WithMessage("Temporal workflows are single-threaded; remove lock/Mutex usage"),
		Matches: []string{"new Mutex", "new Semaphore", "new SemaphoreSlim"},
	},
	{
		Rule:    rules.EnvAccess.WithMessage("Move environment access to an activity or use workflow config"),
		Matches: []string{"Environment.GetEnvironmentVariable", "Environment.GetCommandLineArgs"},
	},
	{
		Rule:    rules.StandardLogging.WithMessage("Use Workflow.Logger instead of Console.Write in workflows"),
		Matches: []string{"Console.Write", "Console.WriteLine"},
	},
}
