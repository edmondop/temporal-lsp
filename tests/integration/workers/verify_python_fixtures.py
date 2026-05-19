import asyncio
import sys
import importlib.util
from pathlib import Path
from temporalio.testing import WorkflowEnvironment
from temporalio.worker import Worker, UnsandboxedWorkflowRunner


FIXTURES_DIR = Path("/testdata")


def load_module(path: Path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def get_workflow_classes(mod):
    classes = []
    for name in dir(mod):
        obj = getattr(mod, name)
        if isinstance(obj, type) and hasattr(obj, "__temporal_workflow_definition"):
            classes.append(obj)
    return classes


def get_activity_functions(mod):
    funcs = []
    for name in dir(mod):
        obj = getattr(mod, name)
        if callable(obj) and hasattr(obj, "__temporal_activity_definition"):
            funcs.append(obj)
    return funcs


async def verify_fixture(env: WorkflowEnvironment, fixture_path: Path, execute: bool):
    print(f"  Verifying {fixture_path.name}...")

    mod = load_module(fixture_path)
    workflows = get_workflow_classes(mod)
    activities = get_activity_functions(mod)

    if not workflows and not activities:
        print(f"    FAIL: no workflows or activities found in {fixture_path.name}")
        return False

    task_queue = f"test-{fixture_path.stem}"

    async with Worker(
        env.client,
        task_queue=task_queue,
        workflows=workflows,
        activities=activities,
        workflow_runner=UnsandboxedWorkflowRunner(),
    ):
        print(f"    OK: worker registered ({len(workflows)} workflows, {len(activities)} activities)")

        if execute and workflows:
            wf_class = workflows[0]
            try:
                result = await env.client.execute_workflow(
                    wf_class.run,
                    "test-input",
                    id=f"verify-{fixture_path.stem}",
                    task_queue=task_queue,
                )
                print(f"    OK: workflow executed, result={result}")
            except Exception as e:
                print(f"    FAIL: workflow execution failed: {e}")
                return False

    return True


async def main():
    async with await WorkflowEnvironment.start_local() as env:
        results = {}

        # Only execute good_workflow.py (accepts str input)
        # good_signatures.py expects a dataclass — just verify it registers
        execute_fixtures = [FIXTURES_DIR / "good_workflow.py"]
        register_fixtures = [
            FIXTURES_DIR / "good_signatures.py",
            FIXTURES_DIR / "bad_workflow.py",
            FIXTURES_DIR / "bad_signatures.py",
        ]

        for fixture in execute_fixtures:
            if fixture.exists():
                results[fixture.name] = await verify_fixture(env, fixture, execute=True)

        for fixture in register_fixtures:
            if fixture.exists():
                results[fixture.name] = await verify_fixture(env, fixture, execute=False)

        all_passed = all(results.values())
        for name, ok in results.items():
            print(f"  {'PASS' if ok else 'FAIL'}: {name}")

        if not all_passed:
            sys.exit(1)

        print("\nAll fixtures verified as real Temporal workflows.")


if __name__ == "__main__":
    asyncio.run(main())
