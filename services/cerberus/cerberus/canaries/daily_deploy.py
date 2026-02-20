"""Long-lived daily deploy canary — deploy once per day for 30 days."""

from cerberus.dsl import *


@canary(
    name="daily_deploy_30d",
    description="Deploy once per day for 30 days, verify each time",
    category="long-lived",
    timeout=days(31),
)
class DailyDeploy:
    app = App(
        name="cerberus-longlived",
        components=[
            TerraformComponent(repo="nuonco/terraform-test", branch="main"),
        ],
    )
    installs = [Install(name="long-lived-install")]

    steps = [
        deploy(),
        repeat(n=30, interval=days(1), steps=[
            deploy(),
            verify_status("healthy"),
            verify_deploy_status("succeeded"),
            log("daily deploy {iteration}/30 complete"),
        ]),
    ]
