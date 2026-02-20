"""Config change + redeploy canary — deploy, change config, redeploy, verify."""

from cerberus.dsl import *


@canary(
    name="config_change_redeploy",
    description="Deploy, change config, redeploy, verify",
    category="basic",
)
class ConfigChangeRedeploy:
    app = App(
        name="cerberus-config-test",
        components=[
            HelmComponent(chart="nginx", namespace="default", values={"replicas": 1}),
        ],
    )
    installs = [Install(name="config-install")]

    steps = [
        deploy(),
        verify_status("healthy"),

        # update component config
        update_component("helm", values={"replicas": 3}),
        build(component="helm"),
        deploy(),
        verify_status("healthy"),
    ]
