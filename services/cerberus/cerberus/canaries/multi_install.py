"""Multi-install canary — deploy to 3 installs, verify all healthy."""

from cerberus.dsl import *


@canary(
    name="multi_install_deploy",
    description="3 installs, deploy to all, verify all healthy",
    category="basic",
)
class MultiInstallDeploy:
    app = App(
        name="cerberus-multi",
        components=[
            TerraformComponent(repo="nuonco/terraform-test", branch="main"),
            HelmComponent(chart="nginx", namespace="default"),
        ],
    )
    installs = [
        Install(name="install-us-east"),
        Install(name="install-us-west"),
        Install(name="install-eu-west"),
    ]

    steps = [
        deploy(),
        verify_status("healthy"),
        delay(hours=1),
        deploy(),
        verify_status("healthy"),
    ]
