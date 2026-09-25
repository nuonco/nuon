from __future__ import annotations

import os
from dataclasses import dataclass
from typing import Any

import httpx

from nuon_ext_install_config_history import __version__


class NuonAPIError(RuntimeError):
    pass


@dataclass(frozen=True)
class NuonContext:
    api_url: str
    token: str
    org_id: str
    install_id: str

    @classmethod
    def from_env(
        cls,
        *,
        api_url: str | None = None,
        token: str | None = None,
        org_id: str | None = None,
        install_id: str | None = None,
    ) -> NuonContext:
        return cls(
            api_url=(api_url or os.getenv("NUON_API_URL") or "https://api.nuon.co").rstrip("/"),
            token=token or os.getenv("NUON_API_TOKEN", ""),
            org_id=org_id or os.getenv("NUON_ORG_ID", ""),
            install_id=install_id or os.getenv("NUON_INSTALL_ID", ""),
        )

    def validate(self) -> None:
        missing = [
            name
            for name, value in (
                ("API token", self.token),
                ("organization", self.org_id),
                ("install", self.install_id),
            )
            if not value
        ]
        if missing:
            names = ", ".join(missing)
            raise NuonAPIError(
                f"missing {names}; authenticate and select an org/install with the nuon CLI, "
                "or pass the corresponding options"
            )


class NuonClient:
    def __init__(self, context: NuonContext, *, transport: httpx.BaseTransport | None = None):
        self.context = context
        self.http = httpx.Client(
            base_url=context.api_url,
            headers={
                "Authorization": f"Bearer {context.token}",
                "X-Nuon-Org-ID": context.org_id,
                "Accept": "application/json",
                "User-Agent": f"nuon-ext-install-config-history/{__version__}",
            },
            timeout=30,
            transport=transport,
        )

    def close(self) -> None:
        self.http.close()

    def __enter__(self) -> NuonClient:
        return self

    def __exit__(self, *_: object) -> None:
        self.close()

    def get(self, path: str, params: dict[str, Any] | None = None) -> Any:
        try:
            response = self.http.get(path, params=params)
            response.raise_for_status()
            return response.json()
        except httpx.HTTPStatusError as exc:
            detail = self._error_detail(exc.response)
            raise NuonAPIError(f"Nuon API returned {exc.response.status_code}: {detail}") from exc
        except (httpx.HTTPError, ValueError) as exc:
            raise NuonAPIError(f"unable to read from Nuon API: {exc}") from exc

    @staticmethod
    def _error_detail(response: httpx.Response) -> str:
        try:
            body = response.json()
        except ValueError:
            return response.text or response.reason_phrase
        if isinstance(body, dict):
            for key in ("description", "message", "error"):
                if body.get(key):
                    return str(body[key])
        return str(body)

    def install(self) -> dict[str, Any]:
        return self.get(f"/v1/installs/{self.context.install_id}")

    def config_history(self, app_id: str, branch_id: str | None) -> list[dict[str, Any]]:
        if branch_id:
            path = f"/v1/apps/{app_id}/branches/{branch_id}/configs"
        else:
            path = f"/v1/apps/{app_id}/configs"
        return self._offset_pages(path)

    def branch_runs(self, app_id: str, branch_id: str | None) -> list[dict[str, Any]]:
        if not branch_id:
            return []
        return self._offset_pages(f"/v1/apps/{app_id}/branches/{branch_id}/runs")

    def install_versions(self) -> list[dict[str, Any]]:
        return self.get(f"/v1/installs/{self.context.install_id}/app-config-versions")

    def app_config_diff(
        self, app_id: str, old_config_id: str, new_config_id: str
    ) -> dict[str, Any]:
        return self.get(
            f"/v1/apps/{app_id}/configs/{new_config_id}/diff",
            {"old_config_id": old_config_id},
        )

    def install_version_diff(self, version_id: str) -> dict[str, Any]:
        return self.get(
            f"/v1/installs/{self.context.install_id}/app-config-versions/{version_id}/diff"
        )

    def components(
        self, app_id: str, branch_id: str | None, query: str | None = None
    ) -> list[dict[str, Any]]:
        params = {key: value for key, value in (("branch_id", branch_id), ("q", query)) if value}
        return self._offset_pages(f"/v1/apps/{app_id}/components", params=params)

    def component_configs(self, app_id: str, component_id: str) -> list[dict[str, Any]]:
        return self._offset_pages(f"/v1/apps/{app_id}/components/{component_id}/configs")

    def component_builds(self, app_id: str, component_id: str) -> list[dict[str, Any]]:
        return self._offset_pages(f"/v1/apps/{app_id}/components/{component_id}/builds")

    def _offset_pages(
        self,
        path: str,
        *,
        limit: int = 100,
        params: dict[str, Any] | None = None,
        max_pages: int = 100,
    ) -> list[dict[str, Any]]:
        items: list[dict[str, Any]] = []
        offset = 0
        for _ in range(max_pages):
            page_params = dict(params or {})
            page_params.update({"limit": limit, "offset": offset})
            page = self.http.get(path, params=page_params)
            try:
                page.raise_for_status()
                payload = page.json()
            except httpx.HTTPStatusError as exc:
                detail = self._error_detail(exc.response)
                raise NuonAPIError(
                    f"Nuon API returned {exc.response.status_code}: {detail}"
                ) from exc
            except (httpx.HTTPError, ValueError) as exc:
                raise NuonAPIError(f"unable to read from Nuon API: {exc}") from exc
            if not isinstance(payload, list):
                raise NuonAPIError(f"expected a list from {path}")
            items.extend(payload)
            has_next = page.headers.get("X-Nuon-Page-Next")
            if not payload or has_next == "false" or (has_next is None and len(payload) < limit):
                return items
            offset += len(payload)
        raise NuonAPIError(f"pagination exceeded {max_pages} pages for {path}")
