package java

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

type bannedCall struct {
	Rule    rules.Rule
	Matches []string
}

var bannedCalls = []bannedCall{
	{
		Rule:    rules.TimeNow.WithMessage("Use Workflow.currentTimeMillis() instead of system time in workflows"),
		Matches: []string{"System.currentTimeMillis", "System.nanoTime", "Instant.now", "LocalDateTime.now", "LocalDate.now", "ZonedDateTime.now", "new Date"},
	},
	{
		Rule:    rules.Sleep.WithMessage("Use Workflow.sleep() instead of Thread.sleep() in workflows"),
		Matches: []string{"Thread.sleep"},
	},
	{
		Rule:    rules.Random.WithMessage("Use Workflow.newRandom() instead of Math.random() or Random in workflows"),
		Matches: []string{"Math.random", "new Random", "ThreadLocalRandom.current"},
	},
	{
		Rule:    rules.IO.WithMessage("Move network/IO calls to an activity"),
		Matches: []string{"new File", "Files.", "new FileInputStream", "new FileOutputStream", "URL.openConnection", "HttpClient.", "new Socket"},
	},
	{
		Rule:    rules.Goroutine.WithMessage("Use Async.function() or Async.procedure() instead of raw threads in workflows"),
		Matches: []string{"new Thread", "Executors.", "executor.submit", "executor.execute", "CompletableFuture.supplyAsync", "CompletableFuture.runAsync"},
	},
	{
		Rule:    rules.Mutex.WithMessage("Temporal workflows are single-threaded; remove lock/synchronized usage"),
		Matches: []string{"new ReentrantLock", "new Semaphore", "new CountDownLatch"},
	},
}
