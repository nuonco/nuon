from __future__ import annotations

from typing import Any

from rich.text import Text
from textual import work
from textual.app import App, ComposeResult
from textual.containers import Horizontal, Vertical
from textual.widgets import (
    DataTable,
    Footer,
    Header,
    Input,
    RichLog,
    Static,
    TabbedContent,
    TabPane,
)

from nuon_ext_install_config_history.client import NuonAPIError, NuonClient
from nuon_ext_install_config_history.render import (
    attach_run_summaries,
    branch_id,
    component_history_renderable,
    diff_renderable,
    short_id,
    status_value,
    timestamp,
)


class HistoryApp(App[None]):
    TITLE = "Nuon install config history"
    CSS = """
    #context { height: 3; padding: 1 2; background: $boost; }
    #history-layout, #component-layout { height: 1fr; }
    #configs, #components { width: 44%; height: 1fr; }
    #diff, #component-history { width: 56%; height: 1fr; border-left: solid $primary; }
    #component-search { margin: 0 1; }
    #component-pane { height: 1fr; }
    .error { color: $error; padding: 1 2; }
    """
    BINDINGS = [
        ("q", "quit", "Quit"),
        ("b", "set_baseline", "Set baseline"),
        ("d", "diff_selected", "Diff"),
        ("r", "refresh", "Refresh"),
        ("/", "search", "Search"),
    ]

    def __init__(self, client: NuonClient):
        super().__init__()
        self.client = client
        self.install: dict[str, Any] = {}
        self.configs: list[dict[str, Any]] = []
        self.components: list[dict[str, Any]] = []
        self.visible_components: list[dict[str, Any]] = []
        self.app_id = ""
        self.branch_id: str | None = None
        self.baseline_id: str | None = None
        self._suppress_highlight = False

    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("Loading install context…", id="context")
        with TabbedContent():
            with TabPane("Config history", id="history"):
                with Horizontal(id="history-layout"):
                    yield DataTable(id="configs", cursor_type="row", zebra_stripes=True)
                    yield RichLog(id="diff", wrap=True, markup=True)
            with TabPane("Components", id="component-pane"):
                with Vertical():
                    yield Input(
                        placeholder="Search components by name, type, or ID",
                        id="component-search",
                    )
                    with Horizontal(id="component-layout"):
                        yield DataTable(id="components", cursor_type="row", zebra_stripes=True)
                        yield RichLog(id="component-history", wrap=True, markup=True)
        yield Footer()

    def on_mount(self) -> None:
        self._load_history(notify=False)

    def action_refresh(self) -> None:
        self._load_history(notify=True)

    @work(thread=True, exclusive=True, group="refresh")
    def _load_history(self, notify: bool = False) -> None:
        self.call_from_thread(self._set_loading, True, "Loading install context…")
        try:
            install = self.client.install()
            app_id = str(install.get("app_id") or "")
            app_branch_id = branch_id(install)
            if not app_id:
                raise NuonAPIError("the selected install did not return an app_id")
            configs = self.client.config_history(app_id, app_branch_id)
            try:
                workflows = self.client.branch_runs(app_id, app_branch_id)
            except NuonAPIError:
                workflows = []
            attach_run_summaries(configs, workflows)
            components = self.client.components(app_id, app_branch_id)
        except NuonAPIError as exc:
            self.call_from_thread(self._show_error, str(exc))
            return
        self.call_from_thread(
            self._apply_history, install, app_id, app_branch_id, configs, components, notify
        )

    def _apply_history(
        self,
        install: dict[str, Any],
        app_id: str,
        app_branch_id: str | None,
        configs: list[dict[str, Any]],
        components: list[dict[str, Any]],
        notify: bool,
    ) -> None:
        self.install = install
        self.app_id = app_id
        self.branch_id = app_branch_id
        self.baseline_id = None
        self.configs = configs
        self.components = components
        self.visible_components = components
        self._render_context()
        self._suppress_highlight = True
        self._render_configs()
        self._render_components()
        self._suppress_highlight = False
        self._set_loading(False)
        if self.configs:
            new_id = str(self.configs[0].get("id") or "")
            old_id = str(self.configs[1].get("id") or "") if len(self.configs) > 1 else ""
            if old_id:
                self._show_diff(old_id, new_id)
        if notify:
            self.notify("History refreshed")

    def _set_loading(self, loading: bool, context: str | None = None) -> None:
        for widget_id in ("configs", "components", "diff", "component-history"):
            self.query_one(f"#{widget_id}").loading = loading
        if context:
            self.query_one("#context", Static).update(context)

    def _render_context(self) -> None:
        branch = self.install.get("app_branch") or {}
        branch_name = branch.get("name") if isinstance(branch, dict) else None
        context = (
            f"[bold]{self.install.get('name', self.client.context.install_id)}[/bold]  "
            f"app [cyan]{short_id(self.app_id)}[/cyan]  "
            f"branch [cyan]{branch_name or short_id(self.branch_id) or 'legacy app history'}[/cyan]"
        )
        self.query_one("#context", Static).update(context)

    def _render_configs(self) -> None:
        table = self.query_one("#configs", DataTable)
        table.clear(columns=True)
        table.add_columns("Version", "Config", "Run", "Status", "Created")
        for config in self.configs:
            config_id = str(config.get("id") or "")
            table.add_row(
                str(config.get("version") or "—"),
                short_id(config_id),
                str(config.get("run_summary") or "—"),
                status_value(config.get("status_v2") or config.get("status")),
                timestamp(config.get("created_at")),
                key=config_id,
            )

    def _render_components(self) -> None:
        table = self.query_one("#components", DataTable)
        table.clear(columns=True)
        table.add_columns("Name", "Type", "ID")
        for component in self.visible_components:
            component_id = str(component.get("id") or "")
            table.add_row(
                str(component.get("name") or "—"),
                str(component.get("type") or "—"),
                short_id(component_id),
                key=component_id,
            )

    def on_tabbed_content_tab_activated(self, event: TabbedContent.TabActivated) -> None:
        if event.pane.id != "component-pane":
            return
        table = self.query_one("#components", DataTable)
        if table.row_count == 0 or table.cursor_row < 0:
            return
        row_key = table.coordinate_to_cell_key(table.cursor_coordinate).row_key
        if row_key.value:
            self._show_component(str(row_key.value))
        if event.input.id != "component-search":
            return
        query = event.value.casefold().strip()
        self.visible_components = [
            component
            for component in self.components
            if query
            in " ".join(str(component.get(key) or "") for key in ("name", "type", "id")).casefold()
        ]
        self._render_components()

    def on_data_table_row_highlighted(self, event: DataTable.RowHighlighted) -> None:
        if self._suppress_highlight or event.row_key is None or event.row_key.value is None:
            return
        value = str(event.row_key.value)
        if event.data_table.id == "configs":
            old_id = self.baseline_id or self._previous_config_id(value)
            if old_id:
                self._show_diff(old_id, value)
            else:
                log = self.query_one("#diff", RichLog)
                log.loading = False
                log.clear()
                log.write("No earlier config to compare.")
        elif event.data_table.id == "components":
            if self.query_one(TabbedContent).active != "component-pane":
                return
            self._show_component(value)

    def _selected_config_id(self) -> str | None:
        table = self.query_one("#configs", DataTable)
        if not self.configs or table.cursor_row < 0:
            return None
        row_key = table.coordinate_to_cell_key(table.cursor_coordinate).row_key
        return str(row_key.value)

    def _previous_config_id(self, config_id: str) -> str | None:
        ids = [str(config.get("id") or "") for config in self.configs]
        try:
            index = ids.index(config_id)
        except ValueError:
            return None
        return ids[index + 1] if index + 1 < len(ids) else None

    def _show_diff(self, old_id: str, new_id: str) -> None:
        log = self.query_one("#diff", RichLog)
        log.clear()
        log.write(
            Text(
                f"Comparing {short_id(old_id)} → {short_id(new_id)}",
                style="bold",
            )
        )
        log.loading = True
        self._fetch_diff(old_id, new_id)

    @work(thread=True, exclusive=True, group="pane")
    def _fetch_diff(self, old_id: str, new_id: str) -> None:
        try:
            payload = self.client.app_config_diff(self.app_id, old_id, new_id)
        except NuonAPIError as exc:
            self.call_from_thread(self._render_diff_error, str(exc))
            return
        self.call_from_thread(self._render_diff, old_id, new_id, payload)

    def _render_diff(self, old_id: str, new_id: str, payload: dict[str, Any]) -> None:
        log = self.query_one("#diff", RichLog)
        log.loading = False
        log.clear()
        log.write(
            Text(
                f"Comparing {short_id(old_id)} → {short_id(new_id)}",
                style="bold",
            )
        )
        log.write(diff_renderable(payload))

    def _render_diff_error(self, message: str) -> None:
        log = self.query_one("#diff", RichLog)
        log.loading = False
        log.write(Text(message, style="red"))

    def _show_component(self, component_id: str) -> None:
        component = next(
            (item for item in self.components if item.get("id") == component_id),
            {"id": component_id},
        )
        log = self.query_one("#component-history", RichLog)
        log.clear()
        log.write(f"Loading {component.get('name', component_id)}…")
        log.loading = True
        self._fetch_component(component, component_id)

    @work(thread=True, exclusive=True, group="pane")
    def _fetch_component(self, component: dict[str, Any], component_id: str) -> None:
        try:
            configs = self.client.component_configs(self.app_id, component_id)
            builds = self.client.component_builds(self.app_id, component_id)
        except NuonAPIError as exc:
            self.call_from_thread(self._render_component_error, str(exc))
            return
        self.call_from_thread(self._render_component, component, configs, builds)

    def _render_component(
        self,
        component: dict[str, Any],
        configs: list[dict[str, Any]],
        builds: list[dict[str, Any]],
    ) -> None:
        log = self.query_one("#component-history", RichLog)
        log.loading = False
        log.clear()
        log.write(component_history_renderable(component, configs, builds))

    def _render_component_error(self, message: str) -> None:
        log = self.query_one("#component-history", RichLog)
        log.loading = False
        log.clear()
        log.write(Text(message, style="red"))

    def action_set_baseline(self) -> None:
        selected = self._selected_config_id()
        if selected:
            self.baseline_id = selected
            self.notify(f"Baseline set to {short_id(selected)}")

    def action_diff_selected(self) -> None:
        selected = self._selected_config_id()
        if selected and self.baseline_id and selected != self.baseline_id:
            self._show_diff(self.baseline_id, selected)
        elif not self.baseline_id:
            self.notify("Select a config and press b to set the baseline", severity="warning")
        else:
            self.notify("Baseline and target are the same config", severity="warning")

    def action_search(self) -> None:
        self.query_one(TabbedContent).active = "component-pane"
        self.query_one("#component-search", Input).focus()

    def _show_error(self, message: str) -> None:
        self._set_loading(False)
        self.query_one("#context", Static).update(Text(message, style="bold red"))
        log = self.query_one("#diff", RichLog)
        log.clear()
        log.write(Text(message, style="red"))
