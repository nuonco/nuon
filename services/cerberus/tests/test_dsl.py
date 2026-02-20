"""Tests for the Cerberus DSL — base types, primitives, registry, and runner."""

from datetime import timedelta

from cerberus.dsl.base import (
    App,
    CanaryContext,
    ExternalImageComponent,
    HelmComponent,
    Install,
    TerraformComponent,
)
from cerberus.dsl.primitives import (
    Step,
    api_call,
    build,
    clear_faults,
    cli_run,
    days,
    delay,
    deploy,
    fault,
    hours,
    log,
    minutes,
    repeat,
    verify_deploy_status,
    verify_status,
)
from cerberus.dsl.registry import CanaryDefinition, canary, get_canary, list_canaries, _registry
from cerberus.dsl.runner import Phase, build_execution_plan


# ---------------------------------------------------------------------------
# Primitives
# ---------------------------------------------------------------------------


class TestPrimitives:
    def test_deploy_defaults(self):
        s = deploy()
        assert s.kind == "deploy"
        assert s.params["wait"] is True
        assert s.params["install"] is None

    def test_deploy_with_options(self):
        s = deploy(wait=False, install="us-east-1", plan_only=True)
        assert s.params["wait"] is False
        assert s.params["install"] == "us-east-1"
        assert s.params["plan_only"] is True

    def test_build(self):
        s = build(component="terraform")
        assert s.kind == "build"
        assert s.params["component"] == "terraform"

    def test_delay(self):
        s = delay(hours=6, minutes=30)
        assert s.kind == "delay"
        assert s.params["seconds"] == 6 * 3600 + 30 * 60

    def test_delay_days(self):
        s = delay(days=1)
        assert s.params["seconds"] == 86400

    def test_repeat(self):
        inner = [deploy(), verify_status("healthy")]
        s = repeat(n=5, steps=inner, interval=days(1))
        assert s.kind == "repeat"
        assert s.params["n"] == 5
        assert s.params["interval_seconds"] == 86400
        assert len(s.params["steps"]) == 2

    def test_fault(self):
        s = fault(type="runner_crash", trigger="during_deploy", delay_seconds=30)
        assert s.kind == "fault"
        assert s.params["type"] == "runner_crash"
        assert s.params["trigger"] == "during_deploy"

    def test_verify_status(self):
        s = verify_status("healthy", install="us-east-1")
        assert s.kind == "verify_status"
        assert s.params["expected"] == "healthy"
        assert s.params["install"] == "us-east-1"

    def test_api_call(self):
        s = api_call("GET", "/v1/apps/{app_id}")
        assert s.kind == "api_call"
        assert s.params["method"] == "GET"
        assert s.params["path"] == "/v1/apps/{app_id}"

    def test_cli_run(self):
        s = cli_run("nuon apps list")
        assert s.kind == "cli_run"
        assert s.params["command"] == "nuon apps list"

    def test_log(self):
        s = log("checkpoint", data={"key": "val"})
        assert s.kind == "log"
        assert s.params["message"] == "checkpoint"

    def test_helpers(self):
        assert days(1) == timedelta(days=1)
        assert hours(2) == timedelta(hours=2)
        assert minutes(30) == timedelta(minutes=30)

    def test_step_repr(self):
        s = deploy(wait=False)
        r = repr(s)
        assert "deploy" in r
        assert "wait=False" in r


# ---------------------------------------------------------------------------
# Base types
# ---------------------------------------------------------------------------


class TestBaseTypes:
    def test_app(self):
        app = App(
            name="test-app",
            components=[
                TerraformComponent(repo="org/repo", branch="main"),
                HelmComponent(chart="nginx", namespace="default"),
            ],
        )
        assert app.name == "test-app"
        assert len(app.components) == 2

    def test_install(self):
        inst = Install(name="us-east-1")
        assert inst.name == "us-east-1"

    def test_component_types(self):
        tf = TerraformComponent(repo="org/repo", branch="main", vars={"key": "val"})
        assert tf.name == "terraform"
        assert tf.vars == {"key": "val"}

        helm = HelmComponent(chart="nginx", values={"replicas": 3})
        assert helm.chart == "nginx"

        ext = ExternalImageComponent(image="nginx:latest")
        assert ext.image == "nginx:latest"


# ---------------------------------------------------------------------------
# Context
# ---------------------------------------------------------------------------


class TestCanaryContext:
    def test_resolve_template(self):
        ctx = CanaryContext(
            org_id="org123",
            app_id="app456",
            install_ids={"us-east": "inst789"},
            last_deploy_id="dep999",
            iteration=3,
        )
        assert ctx.resolve_template("{org_id}") == "org123"
        assert ctx.resolve_template("{app_id}") == "app456"
        assert ctx.resolve_template("{install_id}") == "inst789"
        assert ctx.resolve_template("{last_deploy_id}") == "dep999"
        assert ctx.resolve_template("{iteration}") == "3"
        assert ctx.resolve_template("no vars here") == "no vars here"

    def test_resolve_install_ids(self):
        ctx = CanaryContext(install_ids={"a": "id_a", "b": "id_b"})
        assert ctx.resolve_template("{install_ids[0]}") == "id_a"
        assert ctx.resolve_template("{install_ids[1]}") == "id_b"
        assert ctx.resolve_template("{install_ids[a]}") == "id_a"

    def test_serialization(self):
        ctx = CanaryContext(org_id="org1", app_id="app1", install_ids={"x": "y"})
        d = ctx.to_dict()
        ctx2 = CanaryContext.from_dict(d)
        assert ctx2.org_id == "org1"
        assert ctx2.app_id == "app1"
        assert ctx2.install_ids == {"x": "y"}


# ---------------------------------------------------------------------------
# Registry
# ---------------------------------------------------------------------------


class TestRegistry:
    def setup_method(self):
        # Clear registry before each test
        _registry.clear()

    def test_canary_decorator(self):
        @canary(name="test_canary", description="A test", category="test")
        class TestCanary:
            app = App(name="test", components=[])
            installs = [Install(name="i1")]
            steps = [deploy()]

        defn = get_canary("test_canary")
        assert defn.name == "test_canary"
        assert defn.description == "A test"
        assert defn.category == "test"
        assert len(defn.installs) == 1
        assert len(defn.steps) == 1

    def test_canary_requires_app(self):
        try:
            @canary(name="bad")
            class BadCanary:
                steps = []
            assert False, "Should have raised"
        except TypeError:
            pass

    def test_list_canaries(self):
        @canary(name="c1")
        class C1:
            app = App(name="c1")
            steps = []

        @canary(name="c2")
        class C2:
            app = App(name="c2")
            steps = []

        assert len(list_canaries()) == 2

    def test_get_canary_not_found(self):
        try:
            get_canary("nonexistent")
            assert False, "Should have raised"
        except KeyError:
            pass


# ---------------------------------------------------------------------------
# Runner
# ---------------------------------------------------------------------------


class TestRunner:
    def setup_method(self):
        _registry.clear()

    def test_execution_plan_structure(self):
        @canary(name="plan_test")
        class PlanTest:
            app = App(
                name="test",
                components=[TerraformComponent(repo="org/repo")],
            )
            installs = [Install(name="i1"), Install(name="i2")]
            steps = [deploy(), delay(hours=1), verify_status("healthy")]

        defn = get_canary("plan_test")
        plan = build_execution_plan(defn)

        # Setup: create_org, connect_vcs, add_support, create_app, create_install x2
        setup_steps = [s for s in plan if s.phase == Phase.SETUP]
        assert len(setup_steps) == 6  # org + vcs + support + app + 2 installs

        # Steps: deploy, delay, verify
        user_steps = [s for s in plan if s.phase == Phase.STEPS]
        assert len(user_steps) == 3

        # Cleanup: clear_faults, deprovision x2, delete_app, delete_org
        cleanup_steps = [s for s in plan if s.phase == Phase.CLEANUP]
        assert len(cleanup_steps) == 5  # faults + 2 deprovision + delete_app + delete_org

    def test_indices_are_sequential(self):
        @canary(name="idx_test")
        class IdxTest:
            app = App(name="test", components=[])
            installs = [Install(name="i1")]
            steps = [deploy()]

        defn = get_canary("idx_test")
        plan = build_execution_plan(defn)

        for i, step in enumerate(plan):
            assert step.index == i
