package rules

const (
	PythonSDKImportFrom = "from temporalio"
	PythonSDKImport     = "import temporalio"
)

const (
	GoSDKImport   = `"go.temporal.io/sdk/workflow"`
	GoSDKActivity = `"go.temporal.io/sdk/activity"`
)

const (
	JavaSDKImport = "io.temporal"
)

const (
	RustSDKCrate  = "temporal_sdk"
	RustSDKClient = "temporal_client"
	RustSDKCore   = "temporal_sdk_core"
)

const (
	PyDecoratorWorkflowDefn = "workflow.defn"
	PyDecoratorDefn         = "defn"
	PyDecoratorFullDefn     = "temporalio.workflow.defn"
	PyDecoratorWorkflowRun  = "workflow.run"
	PyDecoratorRun          = "run"
	PyDecoratorFullRun      = "temporalio.workflow.run"
)

const (
	JavaAnnotationWorkflowMethod = "WorkflowMethod"
	JavaAnnotationSignalMethod   = "SignalMethod"
	JavaAnnotationQueryMethod    = "QueryMethod"
	JavaAnnotationActivityMethod = "ActivityMethod"
)

const (
	RustAttrWorkflowRun = "workflow_run"
	RustAttrWorkflow    = "workflow"
	RustAttrActivity    = "activity"
)
