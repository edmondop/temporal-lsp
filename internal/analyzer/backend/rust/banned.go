package rust

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

type bannedCall struct {
	Rule    rules.Rule
	Matches []string
}

var bannedCalls = []bannedCall{
	{
		Rule:    rules.TimeNow.WithMessage("Use workflow time utilities instead of SystemTime/Instant in workflows"),
		Matches: []string{"SystemTime::now", "Instant::now", "Utc::now", "Local::now"},
	},
	{
		Rule:    rules.Sleep.WithMessage("Use Temporal's timer API instead of thread::sleep or tokio::time::sleep in workflows"),
		Matches: []string{"thread::sleep", "tokio::time::sleep", "sleep("},
	},
	{
		Rule:    rules.Random.WithMessage("Use workflow random utilities instead of rand in workflows"),
		Matches: []string{"rand::", "thread_rng", "OsRng", "StdRng"},
	},
	{
		Rule:    rules.IO.WithMessage("Move IO operations to an activity"),
		Matches: []string{"std::fs::", "File::open", "File::create", "TcpStream::", "reqwest::", "std::net::"},
	},
	{
		Rule:    rules.Goroutine.WithMessage("Use Temporal's async primitives instead of spawning threads/tasks in workflows"),
		Matches: []string{"thread::spawn", "tokio::spawn", "tokio::task::spawn"},
	},
	{
		Rule:    rules.Mutex.WithMessage("Temporal workflows are single-threaded; remove Mutex/RwLock usage"),
		Matches: []string{"Mutex::new", "RwLock::new"},
	},
}
