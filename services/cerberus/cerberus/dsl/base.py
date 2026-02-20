"""Base types for the Cerberus DSL: App, Install, and Component declarations."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any


# ---------------------------------------------------------------------------
# Component types
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class TerraformComponent:
    repo: str
    branch: str = "main"
    directory: str = "./"
    vars: dict[str, Any] = field(default_factory=dict)
    env_vars: dict[str, str] = field(default_factory=dict)
    name: str = "terraform"


@dataclass(frozen=True)
class HelmComponent:
    chart: str
    namespace: str = "default"
    values: dict[str, Any] = field(default_factory=dict)
    repo: str | None = None
    branch: str = "main"
    name: str = "helm"


@dataclass(frozen=True)
class DockerBuildComponent:
    repo: str
    dockerfile: str = "Dockerfile"
    context: str = "."
    branch: str = "main"
    build_args: dict[str, str] = field(default_factory=dict)
    name: str = "docker_build"


@dataclass(frozen=True)
class KubernetesManifestComponent:
    repo: str
    directory: str = "./"
    branch: str = "main"
    name: str = "kubernetes_manifest"


@dataclass(frozen=True)
class ExternalImageComponent:
    image: str
    tag: str = "latest"
    name: str = "external_image"


@dataclass(frozen=True)
class JobComponent:
    image: str
    command: list[str] = field(default_factory=list)
    args: list[str] = field(default_factory=list)
    env_vars: dict[str, str] = field(default_factory=dict)
    name: str = "job"


ComponentType = (
    TerraformComponent
    | HelmComponent
    | DockerBuildComponent
    | KubernetesManifestComponent
    | ExternalImageComponent
    | JobComponent
)


# ---------------------------------------------------------------------------
# App & Install declarations
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class App:
    name: str
    components: list[ComponentType] = field(default_factory=list)
    description: str = ""


@dataclass(frozen=True)
class Install:
    name: str


# ---------------------------------------------------------------------------
# Canary context — runtime state available during execution
# ---------------------------------------------------------------------------


@dataclass
class CanaryContext:
    """Mutable runtime state populated during a canary run."""

    org_id: str | None = None
    app_id: str | None = None
    install_ids: dict[str, str] = field(default_factory=dict)  # name -> id
    component_ids: dict[str, str] = field(default_factory=dict)  # name -> id

    last_deploy_id: str | None = None
    last_build_id: str | None = None
    last_workflow_id: str | None = None

    iteration: int = 0  # current repeat iteration (1-indexed when inside repeat)

    data: dict[str, Any] = field(default_factory=dict)  # arbitrary user data

    def first_install_id(self) -> str:
        """Return the first install ID, or raise if none exist."""
        if not self.install_ids:
            raise RuntimeError("No installs created yet")
        return next(iter(self.install_ids.values()))

    def resolve_template(self, s: str) -> str:
        """Interpolate {variable} placeholders from context."""
        if "{" not in s:
            return s

        replacements: dict[str, str] = {
            "org_id": self.org_id or "",
            "app_id": self.app_id or "",
            "install_id": self.first_install_id() if self.install_ids else "",
            "last_deploy_id": self.last_deploy_id or "",
            "last_build_id": self.last_build_id or "",
            "last_workflow_id": self.last_workflow_id or "",
            "iteration": str(self.iteration),
        }

        # Add install_ids[N] and install_ids[name]
        for i, (name, iid) in enumerate(self.install_ids.items()):
            replacements[f"install_ids[{i}]"] = iid
            replacements[f"install_ids[{name}]"] = iid

        # Add component_ids[NAME]
        for name, cid in self.component_ids.items():
            replacements[f"component_ids[{name}]"] = cid

        result = s
        for key, val in replacements.items():
            result = result.replace(f"{{{key}}}", val)

        return result

    def to_dict(self) -> dict[str, Any]:
        """Serialize to a dict for passing through Temporal."""
        return {
            "org_id": self.org_id,
            "app_id": self.app_id,
            "install_ids": self.install_ids,
            "component_ids": self.component_ids,
            "last_deploy_id": self.last_deploy_id,
            "last_build_id": self.last_build_id,
            "last_workflow_id": self.last_workflow_id,
            "iteration": self.iteration,
            "data": self.data,
        }

    @classmethod
    def from_dict(cls, d: dict[str, Any]) -> CanaryContext:
        """Deserialize from a dict."""
        return cls(
            org_id=d.get("org_id"),
            app_id=d.get("app_id"),
            install_ids=d.get("install_ids", {}),
            component_ids=d.get("component_ids", {}),
            last_deploy_id=d.get("last_deploy_id"),
            last_build_id=d.get("last_build_id"),
            last_workflow_id=d.get("last_workflow_id"),
            iteration=d.get("iteration", 0),
            data=d.get("data", {}),
        )
