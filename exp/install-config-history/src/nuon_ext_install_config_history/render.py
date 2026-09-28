from __future__ import annotations

import json
from datetime import datetime
from typing import Any

from rich.console import Group
from rich.json import JSON
from rich.panel import Panel
from rich.table import Table
from rich.text import Text


def short_id(value: Any, length: int = 12) -> str:
    text = str(value or "")
    return text if len(text) <= length else f"{text[:length]}…"


def timestamp(value: Any) -> str:
    text = str(value or "")
    if not text:
        return "—"
    try:
        parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
        return parsed.astimezone().strftime("%Y-%m-%d %H:%M:%S")
    except ValueError:
        return text


def status_value(value: Any) -> str:
    if isinstance(value, dict):
        return str(value.get("status") or value.get("state") or "—")
    return str(value or "—")


def branch_id(install: dict[str, Any]) -> str | None:
    value = install.get("app_branch_id")
    if isinstance(value, str):
        return value or None
    if isinstance(value, dict):
        return value.get("String") or value.get("string") or value.get("value")
    branch = install.get("app_branch")
    return branch.get("id") if isinstance(branch, dict) else None


WORKFLOW_TYPE_LABELS = {
    "app_branches_manual_update": "Manual app config update",
    "app_branches_config_repo_update": "Config update",
    "app_branches_component_repo_update": "Component update",
    "app_branch_config_update": "Config update",
}


def branch_run(workflow: dict[str, Any] | None) -> dict[str, Any]:
    if not workflow:
        return {}
    runs = workflow.get("app_branch_runs") or []
    return runs[0] if runs else {}


def runs_by_app_config(workflows: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    mapping: dict[str, dict[str, Any]] = {}
    for workflow in workflows:
        for run in workflow.get("app_branch_runs") or []:
            config_id = run.get("app_config_id")
            if config_id and config_id not in mapping:
                mapping[str(config_id)] = workflow
    return mapping


def _commit_subject(run: dict[str, Any]) -> str:
    commit = run.get("vcs_connection_commit") or {}
    message = str(commit.get("message") or "")
    return message.split("\n", 1)[0].strip()


def _short_sha(run: dict[str, Any]) -> str:
    sha = str(run.get("head_sha") or (run.get("vcs_connection_commit") or {}).get("sha") or "")
    return sha[:7] if sha else ""


def _trigger_label(run: dict[str, Any]) -> str:
    metadata = run.get("metadata") or {}
    pr_number = metadata.get("pr_number")
    if pr_number is None:
        pr_number = run.get("pr_number")
    if metadata.get("trigger") == "tag" and metadata.get("tag"):
        return f"Tag {metadata['tag']}"
    if pr_number is not None and str(pr_number) != "":
        return f"PR #{pr_number}"
    if metadata.get("trigger") == "github_label" and metadata.get("github_label"):
        return f"Label {metadata['github_label']}"
    return ""


def run_summary(workflow: dict[str, Any] | None) -> str:
    if not workflow:
        return ""
    run = branch_run(workflow)
    trigger = _trigger_label(run)
    commit = _commit_subject(run)
    name = str(workflow.get("name") or "")
    if name == "Manual run":
        name = "Run"
    type_label = WORKFLOW_TYPE_LABELS.get(str(workflow.get("type") or ""), "")
    if trigger and commit:
        title = f"{trigger} · {commit}"
    else:
        title = trigger or commit or name or type_label or "Workflow run"
    sha = _short_sha(run)
    if sha and sha.lower() not in title.lower():
        title = f"{title} ({sha})"
    preview = bool(
        workflow.get("plan_only")
        or run.get("plan_only")
        or run.get("run_type") == "git-preview-run"
    )
    if preview:
        title = f"{title} · preview"
    return title


def attach_run_summaries(
    configs: list[dict[str, Any]], workflows: list[dict[str, Any]]
) -> list[dict[str, Any]]:
    mapping = runs_by_app_config(workflows)
    for config in configs:
        workflow = mapping.get(str(config.get("id") or ""))
        if workflow:
            config["run"] = workflow
            config["run_summary"] = run_summary(workflow)
    return configs


def config_table(configs: list[dict[str, Any]]) -> Table:
    table = Table(title="App branch config history")
    table.add_column("Version", justify="right")
    table.add_column("Config")
    table.add_column("Run")
    table.add_column("Status")
    table.add_column("Created")
    table.add_column("CLI")
    for config in configs:
        table.add_row(
            str(config.get("version") or "—"),
            str(config.get("id") or "—"),
            str(config.get("run_summary") or "—"),
            status_value(config.get("status_v2") or config.get("status")),
            timestamp(config.get("created_at")),
            str(config.get("cli_version") or "—"),
        )
    return table


def component_table(components: list[dict[str, Any]]) -> Table:
    table = Table(title="Components")
    table.add_column("Name")
    table.add_column("Type")
    table.add_column("ID")
    for component in components:
        table.add_row(
            str(component.get("name") or "—"),
            str(component.get("type") or "—"),
            str(component.get("id") or "—"),
        )
    return table


def diff_renderable(payload: dict[str, Any]) -> Group:
    summary = payload.get("summary") or {}
    changed = payload.get("changed")
    heading = Text()
    heading.append("Summary  ", style="bold")
    if summary:
        heading.append(
            "  ".join(f"{key}: {value}" for key, value in summary.items()),
            style="cyan",
        )
    else:
        heading.append("No summary returned", style="dim")

    detail: Any
    if changed:
        detail = Text(str(changed))
    elif summary.get("has_changed") is False:
        detail = Text("No changes.", style="dim")
    else:
        detail = JSON(json.dumps(payload.get("diff") or payload))
    return Group(heading, Panel(detail, border_style="cyan"))


def component_history_renderable(
    component: dict[str, Any],
    configs: list[dict[str, Any]],
    builds: list[dict[str, Any]],
) -> Group:
    config_history = Table(title="Config changes", expand=True)
    config_history.add_column("Version", justify="right")
    config_history.add_column("App config")
    config_history.add_column("Checksum")
    config_history.add_column("Created")
    for config in configs:
        config_history.add_row(
            str(config.get("version") or "—"),
            short_id(config.get("app_config_id")),
            short_id(config.get("checksum")),
            timestamp(config.get("created_at")),
        )

    build_history = Table(title="Builds", expand=True)
    build_history.add_column("Build")
    build_history.add_column("Status")
    build_history.add_column("Git ref")
    build_history.add_column("Created")
    for build in builds:
        build_history.add_row(
            short_id(build.get("id")),
            status_value(build.get("status_v2") or build.get("status")),
            str(build.get("git_ref") or "—"),
            timestamp(build.get("created_at")),
        )

    title = f"{component.get('name', 'component')} · {component.get('type', 'unknown')}"
    return Group(Text(title, style="bold cyan"), config_history, build_history)


def json_output(value: Any) -> str:
    return json.dumps(value, indent=2, sort_keys=True, default=str)
