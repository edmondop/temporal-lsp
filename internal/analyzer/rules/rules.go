package rules

type ID string

const (
	NoTimeNow         ID = "temporal/no-time-now"
	NoSleep           ID = "temporal/no-sleep"
	NoRandom          ID = "temporal/no-random"
	NoIO              ID = "temporal/no-io"
	NoGoroutine       ID = "temporal/no-goroutine"
	NoMutex           ID = "temporal/no-mutex"
	NoChannel         ID = "temporal/no-channel"
	NoEnvAccess       ID = "temporal/no-env-access"
	NoStandardLogging ID = "temporal/no-standard-logging"
)

const (
	SinglePayload  ID = "temporal/single-payload"
	PrimitiveParams ID = "temporal/primitive-params"
)

const (
	ActivityTimeoutRequired ID = "temporal/activity-timeout-required"
	UnboundedLoop           ID = "temporal/unbounded-loop"
	NoContextPropagation    ID = "temporal/no-context-propagation"
	NoNakedError            ID = "temporal/no-naked-error"
	SingleReturn            ID = "temporal/single-return"
	NonDeterministic        ID = "temporal/non-deterministic"
)

const (
	SeverityError   = 1
	SeverityWarning = 2
)

type Range struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

type Violation struct {
	RuleID    ID
	Message   string
	Severity  int
	Range     Range
	Reference string
}

type Rule struct {
	ID       ID
	Message  string
	Severity int
	Ref      string
}

func (r Rule) At(rng Range) Violation {
	return Violation{
		RuleID:    r.ID,
		Message:   r.Message,
		Severity:  r.Severity,
		Range:     rng,
		Reference: r.Ref,
	}
}

func (r Rule) WithMessage(msg string) Rule {
	r.Message = msg
	return r
}

const MistakesRef = "https://github.com/jlegrone/100-temporal-mistakes"
const PayloadRef = MistakesRef + "/blob/main/src/using_more_than_one_input_response_payload/"

var (
	TimeNow         = Rule{ID: NoTimeNow, Severity: SeverityError}
	Sleep           = Rule{ID: NoSleep, Severity: SeverityError}
	Random          = Rule{ID: NoRandom, Severity: SeverityError}
	IO              = Rule{ID: NoIO, Severity: SeverityError}
	Goroutine       = Rule{ID: NoGoroutine, Severity: SeverityError}
	Mutex           = Rule{ID: NoMutex, Severity: SeverityError}
	Channel         = Rule{ID: NoChannel, Severity: SeverityError}
	EnvAccess       = Rule{ID: NoEnvAccess, Severity: SeverityError}
	StandardLogging = Rule{ID: NoStandardLogging, Severity: SeverityError}

	SinglePayloadRule   = Rule{ID: SinglePayload, Severity: SeverityWarning, Ref: PayloadRef}
	PrimitiveParamsRule = Rule{ID: PrimitiveParams, Severity: SeverityWarning, Ref: PayloadRef}

	ActivityTimeout = Rule{ID: ActivityTimeoutRequired, Severity: SeverityWarning, Ref: MistakesRef}
	Unbounded       = Rule{ID: UnboundedLoop, Severity: SeverityWarning, Ref: MistakesRef}

	ContextPropagation = Rule{ID: NoContextPropagation, Severity: SeverityError, Ref: MistakesRef}
	NakedError         = Rule{ID: NoNakedError, Severity: SeverityWarning, Ref: MistakesRef}
	SingleReturnRule   = Rule{ID: SingleReturn, Severity: SeverityWarning, Ref: PayloadRef}
	NonDeterminism     = Rule{ID: NonDeterministic, Severity: SeverityError}
)
