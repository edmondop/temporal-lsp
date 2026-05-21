package python

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

type bannedCall struct {
	Rule    rules.Rule
	Matches []string
}

var bannedCalls = []bannedCall{
	{
		Rule:    rules.TimeNow.WithMessage("Use workflow.now() instead of datetime.now() or time.time() in workflows"),
		Matches: []string{"datetime.datetime.now", "datetime.now", "time.time"},
	},
	{
		Rule:    rules.Sleep.WithMessage("Use workflow.sleep() instead of time.sleep() in workflows"),
		Matches: []string{"time.sleep"},
	},
	{
		Rule:    rules.Random.WithMessage("Use workflow.random() instead of random.* in workflows"),
		Matches: []string{"random."},
	},
	{
		Rule:    rules.IO.WithMessage("Move network/IO calls to an activity"),
		Matches: []string{"requests.", "urllib.", "open("},
	},
	{
		Rule:    rules.Goroutine.WithMessage("Use workflow.start_activity() or asyncio tasks managed by the workflow instead of threading"),
		Matches: []string{"threading.Thread", "threading.start", "multiprocessing.Process", "multiprocessing.Pool"},
	},
	{
		Rule:    rules.Mutex.WithMessage("Temporal workflows are single-threaded; remove lock usage"),
		Matches: []string{"threading.Lock", "threading.RLock", "asyncio.Lock"},
	},
	{
		Rule:    rules.Channel.WithMessage("Use workflow signals or workflow queues instead of queue/multiprocessing primitives"),
		Matches: []string{"queue.Queue", "multiprocessing.Queue", "asyncio.Queue"},
	},
	{
		Rule:    rules.EnvAccess.WithMessage("Environment variables are non-deterministic; pass configuration as workflow input"),
		Matches: []string{"os.getenv", "os.environ"},
	},
	{
		Rule:    rules.StandardLogging.WithMessage("Use workflow.logger instead of standard logging (avoids duplicate messages during replay)"),
		Matches: []string{"logging.", "logger.", "print"},
	},
}
