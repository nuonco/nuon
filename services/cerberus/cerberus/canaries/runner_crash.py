"""Runner crash recovery canary — deploy, crash runner, verify recovery."""

from cerberus.dsl import *


@canary(
    name="runner_crash_recovery",
    description="Deploy, crash runner mid-deploy, verify system recovers",
    category="fault-injection",
)
class RunnerCrashRecovery:
    app = App(
        name="cerberus-fault-test",
        components=[
            HelmComponent(chart="nginx", namespace="default"),
        ],
    )
    installs = [Install(name="fault-install")]

    steps = [
        # baseline: normal deploy works
        deploy(),
        verify_status("healthy"),

        # inject fault, then deploy again
        fault(type="runner_crash", trigger="during_deploy", delay_seconds=30),
        deploy(),

        # clear fault, verify recovery
        clear_faults(),
        delay(minutes=5),
        deploy(),
        verify_status("healthy"),
        verify_deploy_status("succeeded"),
    ]
