"""Temporal activities for Cerberus canary runs.

Each activity executes a single step primitive against the Nuon API,
records events to the database, and returns the updated context.
"""

from __future__ import annotations

import asyncio
import logging
import traceback
from datetime import datetime, timezone
from typing import Any

from nanoid import generate as nanoid
from temporalio import activity

from cerberus.db.models import CanaryRun, RunEvent, RunStep
from cerberus.db.session import get_session
from cerberus.dsl.base import (
    App,
    CanaryContext,
    DockerBuildComponent,
    ExternalImageComponent,
    HelmComponent,
    Install,
    JobComponent,
    KubernetesManifestComponent,
    TerraformComponent,
)
from cerberus.nuon.client import NuonClient
from cerberus.nuon.models import (
    ConnectedGithubVCSConfig,
    CreateAppRequest,
    CreateComponentRequest,
    CreateDockerBuildComponentConfigRequest,
    CreateExternalImageComponentConfigRequest,
    CreateHelmComponentConfigRequest,
    CreateInstallDeployRequest,
    CreateInstallRequest,
    CreateJobComponentConfigRequest,
    CreateKubernetesManifestComponentConfigRequest,
    CreateOrgRequest,
    CreateTerraformComponentConfigRequest,
    CreateVCSConnectionRequest,
)

logger = logging.getLogger(__name__)


def _id() -> str:
    return nanoid(size=26)


def _now() -> datetime:
    return datetime.now(timezone.utc)


# ---------------------------------------------------------------------------
# Activity: execute a single step
# ---------------------------------------------------------------------------


@activity.defn
async def execute_step(
    run_id: str,
    step_index: int,
    step_kind: str,
    step_params: dict[str, Any],
    step_name: str,
    phase: str,
    context_data: dict[str, Any],
    nuon_base_url: str,
    nuon_api_token: str,
) -> dict[str, Any]:
    """Execute a single canary step and return updated context."""
    ctx = CanaryContext.from_dict(context_data)
    client = NuonClient(base_url=nuon_base_url, api_token=nuon_api_token, org_id=ctx.org_id)

    step_id = _id()
    started_at = _now()

    # Record step start
    async with get_session() as db:
        db_step = RunStep(
            id=step_id,
            run_id=run_id,
            step_index=step_index,
            name=step_name,
            phase=phase,
            status="running",
            input_data=step_params,
            started_at=started_at,
        )
        db.add(db_step)

    try:
        output = await _dispatch_step(step_kind, step_params, ctx, client)

        completed_at = _now()
        duration_ms = int((completed_at - started_at).total_seconds() * 1000)

        async with get_session() as db:
            db_step = await db.get(RunStep, step_id)
            if db_step:
                db_step.status = "passed"
                db_step.output_data = output
                db_step.completed_at = completed_at
                db_step.duration_ms = duration_ms

    except Exception as e:
        completed_at = _now()
        duration_ms = int((completed_at - started_at).total_seconds() * 1000)
        tb = traceback.format_exc()

        async with get_session() as db:
            db_step = await db.get(RunStep, step_id)
            if db_step:
                db_step.status = "failed"
                db_step.error_message = str(e)
                db_step.error_traceback = tb
                db_step.completed_at = completed_at
                db_step.duration_ms = duration_ms

            db.add(RunEvent(
                id=_id(),
                run_id=run_id,
                step_id=step_id,
                event_type="error",
                message=f"Step '{step_name}' failed: {e}",
                data={"traceback": tb},
                created_at=completed_at,
            ))

        await client.close()
        raise

    await client.close()
    return ctx.to_dict()


# ---------------------------------------------------------------------------
# Activity: update run status
# ---------------------------------------------------------------------------


@activity.defn
async def update_run_status(
    run_id: str,
    status: str,
    error_message: str | None = None,
    error_step: str | None = None,
) -> None:
    """Update the canary run status in the database."""
    async with get_session() as db:
        run = await db.get(CanaryRun, run_id)
        if run:
            run.status = status
            if error_message:
                run.error_message = error_message
            if error_step:
                run.error_step = error_step
            if status == "running" and run.started_at is None:
                run.started_at = _now()
            if status in ("passed", "failed", "cancelled", "timed_out"):
                run.completed_at = _now()

        db.add(RunEvent(
            id=_id(),
            run_id=run_id,
            event_type="status_change",
            message=f"Run status → {status}",
            data={"status": status, "error_message": error_message},
            created_at=_now(),
        ))


# ---------------------------------------------------------------------------
# Activity: save context (for tracking resources created)
# ---------------------------------------------------------------------------


@activity.defn
async def save_run_context(run_id: str, context_data: dict[str, Any]) -> None:
    """Persist the current context (org_id, app_id, install_ids) to the run record."""
    ctx = CanaryContext.from_dict(context_data)
    async with get_session() as db:
        run = await db.get(CanaryRun, run_id)
        if run:
            run.nuon_org_id = ctx.org_id
            run.nuon_app_id = ctx.app_id
            run.nuon_install_ids = ctx.install_ids


# ---------------------------------------------------------------------------
# Step dispatch — maps step_kind to actual execution
# ---------------------------------------------------------------------------


async def _dispatch_step(
    kind: str,
    params: dict[str, Any],
    ctx: CanaryContext,
    client: NuonClient,
) -> dict[str, Any] | None:
    """Route a step to the appropriate handler."""

    # Setup steps
    if kind == "setup_create_org":
        return await _setup_create_org(params, ctx, client)
    elif kind == "setup_connect_vcs":
        return await _setup_connect_vcs(ctx, client)
    elif kind == "setup_add_support_users":
        return await _setup_add_support_users(ctx, client)
    elif kind == "setup_create_app":
        return await _setup_create_app(params, ctx, client)
    elif kind == "setup_create_install":
        return await _setup_create_install(params, ctx, client)

    # Cleanup steps
    elif kind == "cleanup_deprovision_install":
        return await _cleanup_deprovision_install(params, ctx, client)
    elif kind == "cleanup_delete_app":
        return await _cleanup_delete_app(ctx, client)
    elif kind == "cleanup_delete_org":
        return await _cleanup_delete_org(ctx, client)

    # User step primitives
    elif kind == "deploy":
        return await _step_deploy(params, ctx, client)
    elif kind == "build":
        return await _step_build(params, ctx, client)
    elif kind == "teardown":
        return await _step_teardown(params, ctx, client)
    elif kind == "verify_status":
        return await _step_verify_status(params, ctx, client)
    elif kind == "verify_deploy_status":
        return await _step_verify_deploy_status(params, ctx, client)
    elif kind == "verify_build_status":
        return await _step_verify_build_status(params, ctx, client)
    elif kind == "approve_workflow":
        return await _step_approve_workflow(params, ctx, client)
    elif kind == "cancel_workflow":
        return await _step_cancel_workflow(params, ctx, client)
    elif kind == "wait_for_workflow":
        return await _step_wait_for_workflow(params, ctx, client)
    elif kind == "update_component":
        return await _step_update_component(params, ctx, client)
    elif kind == "reprovision":
        return await _step_reprovision(params, ctx, client)
    elif kind == "deprovision":
        return await _step_deprovision(params, ctx, client)
    elif kind == "create_install":
        return await _step_create_install(params, ctx, client)
    elif kind == "delete_install":
        return await _step_delete_install(params, ctx, client)
    elif kind == "fault":
        return await _step_fault(params, ctx, client)
    elif kind == "clear_faults":
        return await _step_clear_faults(ctx, client)
    elif kind == "clear_fault":
        return await _step_clear_fault(params, ctx, client)
    elif kind == "api_call":
        return await _step_api_call(params, ctx, client)
    elif kind == "cli_run":
        return await _step_cli_run(params, ctx, client)
    elif kind == "shell":
        return await _step_shell(params, ctx, client)
    elif kind == "log":
        return {"message": ctx.resolve_template(params.get("message", ""))}

    else:
        raise ValueError(f"Unknown step kind: {kind}")


# ---------------------------------------------------------------------------
# Setup handlers
# ---------------------------------------------------------------------------


async def _setup_create_org(
    params: dict, ctx: CanaryContext, client: NuonClient
) -> dict:
    org = await client.create_org(CreateOrgRequest(
        name=f"cerberus-{_id()[:8]}",
        sandbox_mode=params.get("sandbox_mode", True),
    ))
    ctx.org_id = org.id
    client.set_org_id(org.id)
    return {"org_id": org.id}


async def _setup_connect_vcs(ctx: CanaryContext, client: NuonClient) -> dict:
    import os
    github_install_id = os.environ.get("GITHUB_INSTALL_ID", "")
    if github_install_id:
        vcs = await client.create_vcs_connection(
            CreateVCSConnectionRequest(github_install_id=github_install_id)
        )
        return {"vcs_connection_id": vcs.id}
    return {"skipped": True, "reason": "GITHUB_INSTALL_ID not set"}


async def _setup_add_support_users(ctx: CanaryContext, client: NuonClient) -> dict:
    await client.add_support_users()
    return {"added": True}


async def _setup_create_app(
    params: dict, ctx: CanaryContext, client: NuonClient
) -> dict:
    app_decl: App = params["app"]
    app = await client.create_app(CreateAppRequest(
        name=app_decl.name,
        description=app_decl.description,
    ))
    ctx.app_id = app.id

    # Create components and their configs
    for comp in app_decl.components:
        component = await client.create_component(
            app.id,
            CreateComponentRequest(name=comp.name, kind=_component_kind(comp)),
        )
        ctx.component_ids[comp.name] = component.id
        await _create_component_config(client, component.id, comp)

    return {"app_id": app.id, "component_ids": ctx.component_ids}


async def _setup_create_install(
    params: dict, ctx: CanaryContext, client: NuonClient
) -> dict:
    install_decl: Install = params["install"]
    install = await client.create_install(
        ctx.app_id,
        CreateInstallRequest(name=install_decl.name),
    )
    ctx.install_ids[install_decl.name] = install.id
    return {"install_id": install.id, "install_name": install_decl.name}


# ---------------------------------------------------------------------------
# Cleanup handlers
# ---------------------------------------------------------------------------


async def _cleanup_deprovision_install(
    params: dict, ctx: CanaryContext, client: NuonClient
) -> dict:
    install_name = params["install_name"]
    install_id = ctx.install_ids.get(install_name)
    if install_id:
        try:
            await client.deprovision_install(install_id)
        except Exception as e:
            logger.warning(f"Failed to deprovision install {install_id}: {e}")
    return {"install_name": install_name, "install_id": install_id}


async def _cleanup_delete_app(ctx: CanaryContext, client: NuonClient) -> dict:
    if ctx.app_id:
        try:
            await client.delete_app(ctx.app_id)
        except Exception as e:
            logger.warning(f"Failed to delete app {ctx.app_id}: {e}")
    return {"app_id": ctx.app_id}


async def _cleanup_delete_org(ctx: CanaryContext, client: NuonClient) -> dict:
    if ctx.org_id:
        try:
            await client.delete_org()
        except Exception as e:
            logger.warning(f"Failed to delete org {ctx.org_id}: {e}")
    return {"org_id": ctx.org_id}


# ---------------------------------------------------------------------------
# User step handlers
# ---------------------------------------------------------------------------


async def _step_deploy(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    target_install = params.get("install")
    plan_only = params.get("plan_only", False)
    wait = params.get("wait", True)

    install_ids = (
        [ctx.install_ids[target_install]] if target_install else list(ctx.install_ids.values())
    )

    results = []
    for iid in install_ids:
        if plan_only:
            await client.deploy_install_components(iid, plan_only=True)
            results.append({"install_id": iid, "plan_only": True})
        else:
            dep = await client.create_install_deploy(iid)
            ctx.last_deploy_id = dep.id
            results.append({"install_id": iid, "deploy_id": dep.id})

            if wait:
                await _poll_deploy(client, iid, dep.id, timeout_seconds=600)

    return {"deploys": results}


async def _step_build(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    target_component = params.get("component")
    component_ids = (
        [ctx.component_ids[target_component]] if target_component else list(ctx.component_ids.values())
    )

    results = []
    for cid in component_ids:
        bld = await client.create_component_build(cid)
        ctx.last_build_id = bld.id
        results.append({"component_id": cid, "build_id": bld.id})

    return {"builds": results}


async def _step_teardown(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    target_component = params.get("component")
    for iid in ctx.install_ids.values():
        if target_component:
            cid = ctx.component_ids[target_component]
            await client.teardown_install_component(iid, cid)
        else:
            await client.teardown_install_components(iid)
    return {"teardown": True}


async def _step_verify_status(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    expected = params["expected"]
    target_install = params.get("install")

    install_ids = (
        [ctx.install_ids[target_install]] if target_install else list(ctx.install_ids.values())
    )

    for iid in install_ids:
        install = await client.get_install(iid)
        if install.status != expected:
            raise AssertionError(
                f"Install {iid} status is '{install.status}', expected '{expected}'"
            )

    return {"verified": True, "expected": expected}


async def _step_verify_deploy_status(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    expected = params["expected"]
    if not ctx.last_deploy_id:
        raise AssertionError("No deploy to verify — last_deploy_id is None")

    install_id = next(iter(ctx.install_ids.values()))
    dep = await client.get_install_deploy(install_id, ctx.last_deploy_id)
    if dep.status != expected:
        raise AssertionError(
            f"Deploy {dep.id} status is '{dep.status}', expected '{expected}'"
        )

    return {"deploy_id": dep.id, "status": dep.status, "expected": expected}


async def _step_verify_build_status(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    expected = params["expected"]
    component = params.get("component")

    if component:
        cid = ctx.component_ids[component]
        bld = await client.get_component_latest_build(cid)
    elif ctx.last_build_id:
        bld = await client.get_build(ctx.last_build_id)
    else:
        raise AssertionError("No build to verify")

    if bld.status != expected:
        raise AssertionError(f"Build {bld.id} status is '{bld.status}', expected '{expected}'")

    return {"build_id": bld.id, "status": bld.status}


async def _step_approve_workflow(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    wf_id = ctx.resolve_template(params["workflow_id"])
    steps = await client.get_workflow_steps(wf_id)
    for step in steps:
        if step.status == "awaiting_approval":
            await client.approve_workflow_step(wf_id, step.id)
            return {"workflow_id": wf_id, "step_id": step.id, "approved": True}
    raise AssertionError(f"No step awaiting approval in workflow {wf_id}")


async def _step_cancel_workflow(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    wf_id = ctx.resolve_template(params["workflow_id"])
    await client.cancel_workflow(wf_id)
    return {"workflow_id": wf_id, "cancelled": True}


async def _step_wait_for_workflow(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    expected_status = params["status"]
    timeout_seconds = params.get("timeout_seconds", 1800)

    install_id = next(iter(ctx.install_ids.values()))
    elapsed = 0
    poll_interval = 10

    while elapsed < timeout_seconds:
        workflows = await client.get_workflows(install_id)
        if workflows:
            latest = workflows[0]
            ctx.last_workflow_id = latest.id
            if latest.status == expected_status:
                return {"workflow_id": latest.id, "status": latest.status}
        await asyncio.sleep(poll_interval)
        elapsed += poll_interval

    raise TimeoutError(
        f"Workflow did not reach status '{expected_status}' within {timeout_seconds}s"
    )


async def _step_update_component(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    component_name = params["component"]
    cid = ctx.component_ids[component_name]
    # Re-create config with new params (the API replaces the config)
    # This is a simplified version — in practice you'd read current config and merge
    return {"component_id": cid, "updated": True, "params": params}


async def _step_reprovision(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    target = params.get("install")
    install_ids = (
        [ctx.install_ids[target]] if target else list(ctx.install_ids.values())
    )
    for iid in install_ids:
        await client.reprovision_install(iid)
    return {"reprovisioned": [iid for iid in install_ids]}


async def _step_deprovision(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    target = params.get("install")
    install_ids = (
        [ctx.install_ids[target]] if target else list(ctx.install_ids.values())
    )
    for iid in install_ids:
        await client.deprovision_install(iid)
    return {"deprovisioned": [iid for iid in install_ids]}


async def _step_create_install(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    name = params["name"]
    install = await client.create_install(ctx.app_id, CreateInstallRequest(name=name))
    ctx.install_ids[name] = install.id
    return {"install_id": install.id, "name": name}


async def _step_delete_install(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    name = params["install"]
    iid = ctx.install_ids.pop(name, None)
    if iid:
        await client.delete_install(iid)
    return {"install_id": iid, "name": name}


async def _step_fault(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    # Fault injection — placeholder for now. Will be implemented as
    # signals to the runner or API-level fault injection.
    fault_type = params["type"]
    logger.info(f"Injecting fault: {fault_type} with params {params}")
    return {"fault_type": fault_type, "injected": True, "params": params}


async def _step_clear_faults(ctx: CanaryContext, client: NuonClient) -> dict:
    logger.info("Clearing all faults")
    return {"cleared": True}


async def _step_clear_fault(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    fault_type = params["type"]
    logger.info(f"Clearing fault: {fault_type}")
    return {"fault_type": fault_type, "cleared": True}


async def _step_api_call(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    method = params["method"]
    path = ctx.resolve_template(params["path"])
    body = params.get("body")
    result = await client.raw_request(method, path, body=body)
    return {"method": method, "path": path, "response": result}


async def _step_cli_run(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    command = ctx.resolve_template(params["command"])
    proc = await asyncio.create_subprocess_shell(
        command,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )
    stdout, stderr = await proc.communicate()
    return {
        "command": command,
        "exit_code": proc.returncode,
        "stdout": stdout.decode()[:10000],
        "stderr": stderr.decode()[:10000],
    }


async def _step_shell(params: dict, ctx: CanaryContext, client: NuonClient) -> dict:
    command = ctx.resolve_template(params["command"])
    proc = await asyncio.create_subprocess_shell(
        command,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE,
    )
    stdout, stderr = await proc.communicate()
    return {
        "command": command,
        "exit_code": proc.returncode,
        "stdout": stdout.decode()[:10000],
        "stderr": stderr.decode()[:10000],
    }


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _component_kind(comp) -> str:
    """Map a component dataclass to the API kind string."""
    kind_map = {
        TerraformComponent: "terraform_module",
        HelmComponent: "helm_chart",
        DockerBuildComponent: "docker_build",
        KubernetesManifestComponent: "kubernetes_manifest",
        ExternalImageComponent: "external_image",
        JobComponent: "job",
    }
    return kind_map[type(comp)]


async def _create_component_config(client: NuonClient, component_id: str, comp) -> None:
    """Create the component-specific config based on type."""
    if isinstance(comp, TerraformComponent):
        await client.create_terraform_config(
            component_id,
            CreateTerraformComponentConfigRequest(
                connected_github_vcs_config=ConnectedGithubVCSConfig(
                    repo=comp.repo, branch=comp.branch, directory=comp.directory
                ),
                var=comp.vars,
                env_var=comp.env_vars,
            ),
        )
    elif isinstance(comp, HelmComponent):
        await client.create_helm_config(
            component_id,
            CreateHelmComponentConfigRequest(
                chart_name=comp.chart,
                namespace=comp.namespace,
                values=comp.values,
                connected_github_vcs_config=(
                    ConnectedGithubVCSConfig(repo=comp.repo, branch=comp.branch)
                    if comp.repo
                    else None
                ),
            ),
        )
    elif isinstance(comp, DockerBuildComponent):
        await client.create_docker_build_config(
            component_id,
            CreateDockerBuildComponentConfigRequest(
                dockerfile=comp.dockerfile,
                build_context=comp.context,
                connected_github_vcs_config=ConnectedGithubVCSConfig(
                    repo=comp.repo, branch=comp.branch
                ),
                build_arg=comp.build_args,
            ),
        )
    elif isinstance(comp, KubernetesManifestComponent):
        await client.create_kubernetes_manifest_config(
            component_id,
            CreateKubernetesManifestComponentConfigRequest(
                connected_github_vcs_config=ConnectedGithubVCSConfig(
                    repo=comp.repo, branch=comp.branch, directory=comp.directory
                ),
            ),
        )
    elif isinstance(comp, ExternalImageComponent):
        await client.create_external_image_config(
            component_id,
            CreateExternalImageComponentConfigRequest(
                image_url=comp.image, tag=comp.tag
            ),
        )
    elif isinstance(comp, JobComponent):
        await client.create_job_config(
            component_id,
            CreateJobComponentConfigRequest(
                image_url=comp.image,
                cmd=comp.command,
                args=comp.args,
                env_var=comp.env_vars,
            ),
        )


async def _poll_deploy(
    client: NuonClient,
    install_id: str,
    deploy_id: str,
    timeout_seconds: int = 600,
    poll_interval: int = 10,
) -> None:
    """Poll a deploy until it reaches a terminal status."""
    elapsed = 0
    terminal_statuses = {"succeeded", "failed", "cancelled", "timed_out"}

    while elapsed < timeout_seconds:
        dep = await client.get_install_deploy(install_id, deploy_id)
        if dep.status in terminal_statuses:
            return
        await asyncio.sleep(poll_interval)
        elapsed += poll_interval

    raise TimeoutError(f"Deploy {deploy_id} did not complete within {timeout_seconds}s")
