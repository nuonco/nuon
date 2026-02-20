"""CLI + API smoke test canary — exercise CLI and API endpoints."""

from cerberus.dsl import *


@canary(
    name="cli_smoke_test",
    description="Exercise CLI and API endpoints",
    category="smoke-test",
)
class CLISmokeTest:
    app = App(
        name="cerberus-cli-test",
        components=[
            ExternalImageComponent(image="nginx:latest"),
        ],
    )
    installs = [Install(name="cli-install")]

    steps = [
        cli_run("nuon apps list"),
        cli_run("nuon installs list --app {app_id}"),
        api_call("GET", "/v1/apps/{app_id}"),
        api_call("GET", "/v1/installs/{install_id}"),
        deploy(),
        cli_run("nuon installs deploys list --install {install_id}"),
        verify_deploy_status("succeeded"),
    ]
