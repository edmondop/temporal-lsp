package typescript

import "github.com/edmondop/temporal-lsp/internal/analyzer/rules"

type bannedCall struct {
	Rule    rules.Rule
	Matches []string
}

var bannedCalls = []bannedCall{
	{
		Rule:    rules.TimeNow.WithMessage("Use workflow.now() instead of Date.now() or new Date() in workflows"),
		Matches: []string{"Date.now", "new Date"},
	},
	{
		Rule:    rules.Sleep.WithMessage("Use workflow.sleep() instead of setTimeout in workflows"),
		Matches: []string{"setTimeout", "setInterval"},
	},
	{
		Rule:    rules.Random.WithMessage("Use workflow.random() instead of Math.random() in workflows"),
		Matches: []string{"Math.random"},
	},
	{
		Rule:    rules.IO.WithMessage("Move IO/network calls to an activity"),
		Matches: []string{"fetch", "axios.", "fs.read", "fs.write", "fs.open", "fs.unlink"},
	},
	{
		Rule:    rules.EnvAccess.WithMessage("Move environment access to an activity or workflow config"),
		Matches: []string{"process.env"},
	},
	{
		Rule:    rules.StandardLogging.WithMessage("Use workflow.log instead of console.log in workflows"),
		Matches: []string{"console.log", "console.warn", "console.error", "console.info"},
	},
}
