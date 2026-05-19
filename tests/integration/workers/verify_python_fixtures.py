"""Verify Python fixtures are real Temporal workflows that can register and execute.

This script:
1. Connects to a running Temporal server
2. Registers each fixture's workflows/activities as workers
3. For 'good' fixtures: executes the workflow and verifies success
4. For 'bad' fixtures: verifies they at least register as valid workers

Usage: python verify_python_fixtures.py <temporal_address>
"""

import asyncio
import sys
import importlib.util
from pathlib import Path
from temporalio.client import Client
from temporalio.worker import Worker


FIXTURES_DIR = Path("/testdata")


async def load_module(path: Path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


async def get_workflow_classes(mod):
    classes = []
    for name in dir(mod):
        obj = getattr(mod, name)
        if isinstance(obj, type) and hasattr(obj, "__temporal_workflow_definition"):
            classes.append(obj)
    return classes


async def get_activity_functions(mod):
    funcs = []
    for name in dir(mod):
        obj = getattr(mod, name)
        if callable(obj) and hasattr(obj, "__temporal_activity_definition"):
            funcs.append(obj)
    return funcs


async def verify_fixture(client: Client, fixture_path: Path, expect_execution: bool):
    """Register a fixture as a worker. If expect_execution, run the workflow."""
    print(f"  Verifying {fixture_path.name}...")

    mod = await load_module(fixture_path)
    workflows = await get_workflow_classes(mod)
    activities = await get_activity_functions(mod)

    if not workflows and not activities:
        print(f"    FAIL: no workflows or activities found in {fixture_path.name}")
        return False

    task_queue = f"test-{fixture_path.stem}"

    # Start a worker — this proves the code registers correctly
    async with Worker(
        client,
        task_queue=task_queue,
        workflows=workflows,
        activities=activities,
    ):
        print(f"    OK: worker registered ({len(workflows)} workflows, {len(activities)} activities)")

        if expect_execution and workflows:
            # Execute the first workflow to prove it actually runs
            wf_class = workflows[0]
            try:
                result = await client.execute_workflow(
                    wf_class.run,
                    "test-input",
                    id=f"verify-{fixture_path.stem}",
                    task_queue=task_queue,
                )
                print(f"    OK: workflow executed, result={result}")
            except Exception as e:
                if expect_execution:
                    print(f"    FAIL: workflow execution failed: {e}")
                    return False

    return True


async def main():
    if len(sys.argv) < 2:
        print("Usage: verify_python_fixtures.py <temporal_address>")
        sys.exit(1)

    address = sys.argv[1]
    client = await Client.connect(address)

    results = {}

    # Good fixtures should both register AND execute successfully
    good_fixtures = [
        FIXTURES_DIR / "good_workflow.py",
        FIXTURES_DIR / "good_signatures.py",
    ]

    # Bad fixtures should at least register as valid workers
    bad_fixtures = [
        FIXTURES_DIR / "bad_workflow.py",
        FIXTURES_DIR / "bad_signatures.py",
    ]

    for fixture in good_fixtures:
        if fixture.exists():
            ok = await verify_fixture(client, fixture, expect_execution=True)
            results[fixture.name] = ok

    for fixture in bad_fixtures:
        if fixture.exists():
            ok = await verify_fixture(client, fixture, expect_execution=False)
            results[fixture.name] = ok

    # Report
    print("\n=== Results ===")
    all_passed = True
    for name, ok in results.items():
        status = "PASS" if ok else "FAIL"
        print(f"  {status}: {name}")
        if not ok:
            all_passed = False

    if not all_passed:
        print("\nFAILED: some fixtures are not valid Temporal code")
        sys.exit(1)

    print("\nAll fixtures verified as real Temporal workflows.")


if __name__ == "__main__":
    asyncio.run(main())
