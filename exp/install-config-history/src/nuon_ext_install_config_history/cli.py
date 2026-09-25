from __future__ import annotations

import sys
from typing import Any

import click
from rich.console import Console

from nuon_ext_install_config_history.client import NuonAPIError, NuonClient, NuonContext
from nuon_ext_install_config_history.render import (
    attach_run_summaries,
    branch_id,
    component_history_renderable,
    component_table,
    config_table,
    diff_renderable,
    json_output,
)
from nuon_ext_install_config_history.tui import HistoryApp

console = Console()


def resolve_install(client: NuonClient) -> tuple[dict[str, Any], str, str | None]:
    install = client.install()
    app_id = str(install.get("app_id") or "")
    if not app_id:
        raise NuonAPIError("the selected install did not return an app_id")
    return install, app_id, branch_id(install)


def emit(value: Any, output: str, renderable: Any) -> None:
    if output == "json":
        click.echo(json_output(value))
    else:
        console.print(renderable)


def client_for(context: NuonContext) -> NuonClient:
    context.validate()
    return NuonClient(context)


@click.group(invoke_without_command=True, context_settings={"max_content_width": 110})
@click.version_option(package_name="nuon-ext-install-config-history")
@click.option("--install-id", envvar="NUON_INSTALL_ID", metavar="ID", help="Install ID.")
@click.option("--org-id", envvar="NUON_ORG_ID", metavar="ID", help="Organization ID.")
@click.option(
    "--api-url",
    envvar="NUON_API_URL",
    default="https://api.nuon.co",
    show_default=True,
    help="Nuon API base URL.",
)
@click.pass_context
def main(
    ctx: click.Context,
    install_id: str | None,
    org_id: str | None,
    api_url: str,
) -> None:
    """Navigate an install's app branch config history.

    Without a subcommand, opens an interactive browser. Select a config to diff
    it against the preceding config, or press b on one config and d on another
    to compare arbitrary revisions. The Components tab searches components and
    shows each component's config-change and build histories.
    """
    context = NuonContext.from_env(
        api_url=api_url,
        org_id=org_id,
        install_id=install_id,
    )
    ctx.obj = context

    if ctx.invoked_subcommand is None:
        if not sys.stdin.isatty() or not sys.stdout.isatty():
            raise click.ClickException(
                "interactive terminal required; use history, diff, or component"
            )
        try:
            with client_for(context) as client:
                HistoryApp(client).run()
        except NuonAPIError as exc:
            raise click.ClickException(str(exc)) from exc


@main.command("history")
@click.option(
    "--output",
    type=click.Choice(["table", "json"]),
    default="table",
    show_default=True,
)
@click.pass_obj
def history(context: NuonContext, output: str) -> None:
    """List the selected install's app branch configs."""
    try:
        with client_for(context) as client:
            install, app_id, app_branch_id = resolve_install(client)
            configs = client.config_history(app_id, app_branch_id)
            try:
                workflows = client.branch_runs(app_id, app_branch_id)
            except NuonAPIError:
                workflows = []
            attach_run_summaries(configs, workflows)
            payload = {
                "install_id": install.get("id"),
                "app_id": app_id,
                "app_branch_id": app_branch_id,
                "configs": configs,
            }
            emit(payload, output, config_table(configs))
    except NuonAPIError as exc:
        raise click.ClickException(str(exc)) from exc


@main.command("diff")
@click.argument("old_config_id")
@click.argument("new_config_id")
@click.option(
    "--output",
    type=click.Choice(["table", "json"]),
    default="table",
    show_default=True,
)
@click.pass_obj
def diff(
    context: NuonContext,
    old_config_id: str,
    new_config_id: str,
    output: str,
) -> None:
    """Compare OLD_CONFIG_ID with NEW_CONFIG_ID."""
    try:
        with client_for(context) as client:
            _, app_id, _ = resolve_install(client)
            payload = client.app_config_diff(app_id, old_config_id, new_config_id)
            emit(payload, output, diff_renderable(payload))
    except NuonAPIError as exc:
        raise click.ClickException(str(exc)) from exc


@main.command("component")
@click.argument("query")
@click.option(
    "--output",
    type=click.Choice(["table", "json"]),
    default="table",
    show_default=True,
)
@click.pass_obj
def component(context: NuonContext, query: str, output: str) -> None:
    """Search component names, types, and IDs, then show their histories."""
    try:
        with client_for(context) as client:
            _, app_id, app_branch_id = resolve_install(client)
            components = client.components(app_id, app_branch_id)
            needle = query.casefold()
            matches = [
                item
                for item in components
                if needle
                in " ".join(str(item.get(key) or "") for key in ("name", "type", "id")).casefold()
            ]
            if not matches:
                raise NuonAPIError(f"no component matched {query!r}")

            if len(matches) > 1:
                if output == "json":
                    click.echo(json_output({"matches": matches}))
                else:
                    console.print(component_table(matches))
                    console.print(
                        "[yellow]More than one component matched; "
                        "rerun with an exact name or ID.[/yellow]"
                    )
                return

            selected = matches[0]
            component_id = str(selected["id"])
            configs = client.component_configs(app_id, component_id)
            builds = client.component_builds(app_id, component_id)
            payload = {"component": selected, "configs": configs, "builds": builds}
            emit(
                payload,
                output,
                component_history_renderable(selected, configs, builds),
            )
    except NuonAPIError as exc:
        raise click.ClickException(str(exc)) from exc


if __name__ == "__main__":
    main()
