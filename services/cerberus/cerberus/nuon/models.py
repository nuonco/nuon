"""Pydantic models for Nuon API requests and responses."""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel


# ---------------------------------------------------------------------------
# Request models
# ---------------------------------------------------------------------------


class CreateOrgRequest(BaseModel):
    name: str
    sandbox_mode: bool = True


class CreateAppRequest(BaseModel):
    name: str
    description: str = ""


class CreateComponentRequest(BaseModel):
    name: str
    kind: str  # terraform_module, helm_chart, docker_build, etc.


class CreateTerraformComponentConfigRequest(BaseModel):
    connected_github_vcs_config: ConnectedGithubVCSConfig | None = None
    env_var: dict[str, str] = {}
    var: dict[str, Any] = {}


class CreateHelmComponentConfigRequest(BaseModel):
    chart_name: str
    namespace: str = "default"
    values: dict[str, Any] = {}
    connected_github_vcs_config: ConnectedGithubVCSConfig | None = None


class CreateDockerBuildComponentConfigRequest(BaseModel):
    dockerfile: str = "Dockerfile"
    build_context: str = "."
    connected_github_vcs_config: ConnectedGithubVCSConfig | None = None
    build_arg: dict[str, str] = {}


class CreateKubernetesManifestComponentConfigRequest(BaseModel):
    connected_github_vcs_config: ConnectedGithubVCSConfig | None = None


class CreateExternalImageComponentConfigRequest(BaseModel):
    image_url: str
    tag: str = "latest"


class CreateJobComponentConfigRequest(BaseModel):
    image_url: str
    cmd: list[str] = []
    args: list[str] = []
    env_var: dict[str, str] = {}


class ConnectedGithubVCSConfig(BaseModel):
    repo: str
    branch: str = "main"
    directory: str = "./"


class CreateInstallRequest(BaseModel):
    name: str


class CreateInstallDeployRequest(BaseModel):
    pass


class CreateComponentBuildRequest(BaseModel):
    pass


class CreateVCSConnectionRequest(BaseModel):
    github_install_id: str


class CreateAppSecretRequest(BaseModel):
    name: str
    value: str


class UpdateComponentRequest(BaseModel):
    pass


class UpdateInstallInputsRequest(BaseModel):
    inputs: dict[str, Any] = {}


class CreateWorkflowStepApprovalResponseRequest(BaseModel):
    approved: bool = True


# ---------------------------------------------------------------------------
# Response models (simplified — we only extract what we need)
# ---------------------------------------------------------------------------


class Org(BaseModel):
    id: str
    name: str
    sandbox_mode: bool = False

    class Config:
        extra = "allow"


class AppModel(BaseModel):
    id: str
    name: str

    class Config:
        extra = "allow"


class Component(BaseModel):
    id: str
    name: str

    class Config:
        extra = "allow"


class InstallModel(BaseModel):
    id: str
    name: str
    status: str | None = None

    class Config:
        extra = "allow"


class Deploy(BaseModel):
    id: str
    status: str | None = None

    class Config:
        extra = "allow"


class Build(BaseModel):
    id: str
    status: str | None = None

    class Config:
        extra = "allow"


class VCSConnection(BaseModel):
    id: str

    class Config:
        extra = "allow"


class Workflow(BaseModel):
    id: str
    status: str | None = None

    class Config:
        extra = "allow"


class WorkflowStep(BaseModel):
    id: str
    status: str | None = None

    class Config:
        extra = "allow"


class CanaryUser(BaseModel):
    api_token: str
    github_install_id: str

    class Config:
        extra = "allow"
