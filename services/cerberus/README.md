# Cerberus

*Three-headed guardian of the Nuon platform. Testing. Watching. Protecting.*

Cerberus is Nuon's canary testing system v2 — a Python-based framework for declaring, executing, and debugging end-to-end platform tests using Temporal workflows.

## What It Does

Cerberus creates sandbox orgs, apps, and installs against the Nuon platform, then executes declarative test sequences (deploys, builds, config changes, fault injections) to verify the system behaves correctly. Every action is logged, every result is inspectable, and long-lived tests can run for days or weeks.

## Quick Example

```python
from cerberus.dsl import *

@canary(name="basic_deploy", description="Create app, deploy, verify")
class BasicDeploy:
    app = App(
        name="cerberus-basic",
        components=[
            TerraformComponent(repo="nuonco/terraform-test", branch="main"),
            HelmComponent(chart="nginx", namespace="default"),
        ],
    )
    installs = [Install(name="test-install-1")]

    steps = [
        deploy(),
        verify_status("healthy"),
    ]
```

The framework automatically:
1. Creates a sandbox org + VCS connection + support users
2. Creates the app with declared components
3. Creates installs
4. Executes steps in order (each as a Temporal activity with full event logging)
5. Cleans everything up (even on failure)

## Architecture

- **DSL**: Declarative Python classes with primitives like `deploy()`, `delay(days=1)`, `fault(type="runner_crash")`
- **Temporal**: One workflow per canary run, steps as activities, `workflow.sleep()` for long-lived tests
- **Web UI**: FastAPI + HTMX for live-updating run timelines, step inspection, and fault injection controls
- **Database**: PostgreSQL for run history, step results, and event audit logs

## Development

```bash
# Install dependencies
uv sync

# Run the Temporal worker
python -m cerberus.temporal.worker

# Run the web UI
python -m cerberus.web.app

# Run tests
pytest
```
