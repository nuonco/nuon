"""Basic deploy canary — create app, deploy, verify, cleanup."""

from cerberus.dsl import *


@canary(name="basic_deploy", description="Create app, deploy, verify healthy", category="basic")
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
