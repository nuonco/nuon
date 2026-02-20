"""Async Nuon API client for Cerberus.

Wraps the Nuon REST API using httpx. Modeled after sdks/nuon-go/client.go.
"""

from __future__ import annotations

import logging
from typing import Any

import httpx

from cerberus.nuon.models import (
    AppModel,
    Build,
    CanaryUser,
    Component,
    ConnectedGithubVCSConfig,
    CreateAppRequest,
    CreateAppSecretRequest,
    CreateComponentBuildRequest,
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
    CreateWorkflowStepApprovalResponseRequest,
    Deploy,
    InstallModel,
    Org,
    VCSConnection,
    Workflow,
    WorkflowStep,
)

logger = logging.getLogger(__name__)


class NuonAPIError(Exception):
    def __init__(self, status: int, detail: str):
        self.status = status
        self.detail = detail
        super().__init__(f"Nuon API error {status}: {detail}")


class NuonClient:
    """Async client for the Nuon ctl-api."""

    def __init__(
        self,
        base_url: str,
        api_token: str,
        org_id: str | None = None,
    ):
        self.base_url = base_url.rstrip("/")
        self.org_id = org_id
        self._http = httpx.AsyncClient(
            base_url=self.base_url,
            headers=self._build_headers(api_token, org_id),
            timeout=60.0,
        )

    @staticmethod
    def _build_headers(api_token: str, org_id: str | None = None) -> dict[str, str]:
        headers = {
            "Authorization": f"Bearer {api_token}",
            "Content-Type": "application/json",
        }
        if org_id:
            headers["X-Nuon-Org-ID"] = org_id
        return headers

    def set_org_id(self, org_id: str) -> None:
        self.org_id = org_id
        self._http.headers["X-Nuon-Org-ID"] = org_id

    async def close(self) -> None:
        await self._http.aclose()

    # -----------------------------------------------------------------------
    # Internal helpers
    # -----------------------------------------------------------------------

    async def _request(
        self,
        method: str,
        path: str,
        *,
        json: Any = None,
        params: dict | None = None,
    ) -> Any:
        resp = await self._http.request(method, path, json=json, params=params)
        if resp.status_code >= 400:
            detail = resp.text[:500]
            raise NuonAPIError(resp.status_code, detail)
        if resp.status_code == 204:
            return None
        return resp.json()

    async def _get(self, path: str, **kwargs) -> Any:
        return await self._request("GET", path, **kwargs)

    async def _post(self, path: str, **kwargs) -> Any:
        return await self._request("POST", path, **kwargs)

    async def _put(self, path: str, **kwargs) -> Any:
        return await self._request("PUT", path, **kwargs)

    async def _delete(self, path: str, **kwargs) -> Any:
        return await self._request("DELETE", path, **kwargs)

    # -----------------------------------------------------------------------
    # Canary user (admin endpoint)
    # -----------------------------------------------------------------------

    @classmethod
    async def create_canary_client(
        cls,
        admin_base_url: str,
        admin_token: str,
    ) -> NuonClient:
        """Create a new canary user and return a client authenticated as that user."""
        async with httpx.AsyncClient(
            base_url=admin_base_url.rstrip("/"),
            headers={"Authorization": f"Bearer {admin_token}"},
            timeout=30.0,
        ) as http:
            resp = await http.post("/v1/general/canary-user")
            if resp.status_code >= 400:
                raise NuonAPIError(resp.status_code, resp.text[:500])
            data = resp.json()

        user = CanaryUser.model_validate(data)
        return cls(base_url=admin_base_url, api_token=user.api_token)

    # -----------------------------------------------------------------------
    # Orgs
    # -----------------------------------------------------------------------

    async def create_org(self, req: CreateOrgRequest) -> Org:
        data = await self._post("/v1/orgs", json=req.model_dump())
        return Org.model_validate(data)

    async def get_org(self) -> Org:
        data = await self._get("/v1/orgs/current")
        return Org.model_validate(data)

    async def delete_org(self) -> None:
        await self._delete("/v1/orgs/current")

    async def add_support_users(self) -> None:
        await self._post(f"/v1/orgs/{self.org_id}/admin-support-users")

    # -----------------------------------------------------------------------
    # VCS
    # -----------------------------------------------------------------------

    async def create_vcs_connection(self, req: CreateVCSConnectionRequest) -> VCSConnection:
        data = await self._post("/v1/vcs-connections", json=req.model_dump())
        return VCSConnection.model_validate(data)

    # -----------------------------------------------------------------------
    # Apps
    # -----------------------------------------------------------------------

    async def create_app(self, req: CreateAppRequest) -> AppModel:
        data = await self._post("/v1/apps", json=req.model_dump())
        return AppModel.model_validate(data)

    async def get_app(self, app_id: str) -> AppModel:
        data = await self._get(f"/v1/apps/{app_id}")
        return AppModel.model_validate(data)

    async def delete_app(self, app_id: str) -> None:
        await self._delete(f"/v1/apps/{app_id}")

    # -----------------------------------------------------------------------
    # Components
    # -----------------------------------------------------------------------

    async def create_component(self, app_id: str, req: CreateComponentRequest) -> Component:
        data = await self._post(f"/v1/apps/{app_id}/components", json=req.model_dump())
        return Component.model_validate(data)

    async def get_app_components(self, app_id: str) -> list[Component]:
        data = await self._get(f"/v1/apps/{app_id}/components")
        return [Component.model_validate(c) for c in data]

    async def create_terraform_config(
        self, component_id: str, req: CreateTerraformComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/terraform-module",
            json=req.model_dump(),
        )

    async def create_helm_config(
        self, component_id: str, req: CreateHelmComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/helm-chart",
            json=req.model_dump(),
        )

    async def create_docker_build_config(
        self, component_id: str, req: CreateDockerBuildComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/docker-build",
            json=req.model_dump(),
        )

    async def create_kubernetes_manifest_config(
        self, component_id: str, req: CreateKubernetesManifestComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/kubernetes-manifest",
            json=req.model_dump(),
        )

    async def create_external_image_config(
        self, component_id: str, req: CreateExternalImageComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/external-image",
            json=req.model_dump(),
        )

    async def create_job_config(
        self, component_id: str, req: CreateJobComponentConfigRequest
    ) -> dict:
        return await self._post(
            f"/v1/components/{component_id}/config/job",
            json=req.model_dump(),
        )

    # -----------------------------------------------------------------------
    # Builds
    # -----------------------------------------------------------------------

    async def create_component_build(
        self, component_id: str, req: CreateComponentBuildRequest | None = None
    ) -> Build:
        data = await self._post(
            f"/v1/components/{component_id}/builds",
            json=(req or CreateComponentBuildRequest()).model_dump(),
        )
        return Build.model_validate(data)

    async def get_build(self, build_id: str) -> Build:
        data = await self._get(f"/v1/builds/{build_id}")
        return Build.model_validate(data)

    async def get_component_latest_build(self, component_id: str) -> Build:
        data = await self._get(f"/v1/components/{component_id}/builds/latest")
        return Build.model_validate(data)

    # -----------------------------------------------------------------------
    # Installs
    # -----------------------------------------------------------------------

    async def create_install(self, app_id: str, req: CreateInstallRequest) -> InstallModel:
        data = await self._post(f"/v1/apps/{app_id}/installs", json=req.model_dump())
        return InstallModel.model_validate(data)

    async def get_install(self, install_id: str) -> InstallModel:
        data = await self._get(f"/v1/installs/{install_id}")
        return InstallModel.model_validate(data)

    async def deprovision_install(self, install_id: str) -> None:
        await self._post(f"/v1/installs/{install_id}/deprovision")

    async def reprovision_install(self, install_id: str) -> None:
        await self._post(f"/v1/installs/{install_id}/reprovision")

    async def delete_install(self, install_id: str) -> None:
        await self._delete(f"/v1/installs/{install_id}")

    # -----------------------------------------------------------------------
    # Deploys
    # -----------------------------------------------------------------------

    async def create_install_deploy(
        self, install_id: str, req: CreateInstallDeployRequest | None = None
    ) -> Deploy:
        data = await self._post(
            f"/v1/installs/{install_id}/deploys",
            json=(req or CreateInstallDeployRequest()).model_dump(),
        )
        return Deploy.model_validate(data)

    async def get_install_deploy(self, install_id: str, deploy_id: str) -> Deploy:
        data = await self._get(f"/v1/installs/{install_id}/deploys/{deploy_id}")
        return Deploy.model_validate(data)

    async def get_install_latest_deploy(self, install_id: str) -> Deploy:
        data = await self._get(f"/v1/installs/{install_id}/deploys/latest")
        return Deploy.model_validate(data)

    async def deploy_install_components(self, install_id: str, plan_only: bool = False) -> None:
        await self._post(
            f"/v1/installs/{install_id}/components/deploy",
            params={"plan_only": str(plan_only).lower()},
        )

    async def teardown_install_components(self, install_id: str) -> None:
        await self._post(f"/v1/installs/{install_id}/components/teardown")

    async def teardown_install_component(self, install_id: str, component_id: str) -> None:
        await self._post(f"/v1/installs/{install_id}/components/{component_id}/teardown")

    # -----------------------------------------------------------------------
    # Workflows
    # -----------------------------------------------------------------------

    async def get_workflows(self, install_id: str) -> list[Workflow]:
        data = await self._get(f"/v1/installs/{install_id}/workflows")
        return [Workflow.model_validate(w) for w in data]

    async def get_workflow(self, workflow_id: str) -> Workflow:
        data = await self._get(f"/v1/workflows/{workflow_id}")
        return Workflow.model_validate(data)

    async def cancel_workflow(self, workflow_id: str) -> None:
        await self._post(f"/v1/workflows/{workflow_id}/cancel")

    async def get_workflow_steps(self, workflow_id: str) -> list[WorkflowStep]:
        data = await self._get(f"/v1/workflows/{workflow_id}/steps")
        return [WorkflowStep.model_validate(s) for s in data]

    async def approve_workflow_step(
        self, workflow_id: str, step_id: str
    ) -> None:
        await self._post(
            f"/v1/workflows/{workflow_id}/steps/{step_id}/approval-response",
            json=CreateWorkflowStepApprovalResponseRequest().model_dump(),
        )

    async def retry_workflow_step(self, workflow_id: str, step_id: str) -> None:
        await self._post(f"/v1/workflows/{workflow_id}/steps/{step_id}/retry", json={})

    # -----------------------------------------------------------------------
    # Secrets
    # -----------------------------------------------------------------------

    async def create_app_secret(self, app_id: str, req: CreateAppSecretRequest) -> dict:
        return await self._post(f"/v1/apps/{app_id}/secrets", json=req.model_dump())

    async def delete_app_secret(self, app_id: str, secret_id: str) -> None:
        await self._delete(f"/v1/apps/{app_id}/secrets/{secret_id}")

    # -----------------------------------------------------------------------
    # Install inputs
    # -----------------------------------------------------------------------

    async def update_install_inputs(self, install_id: str, inputs: dict[str, Any]) -> dict:
        return await self._put(
            f"/v1/installs/{install_id}/inputs",
            json={"inputs": inputs},
        )

    # -----------------------------------------------------------------------
    # Raw API call (for api_call() primitive)
    # -----------------------------------------------------------------------

    async def raw_request(self, method: str, path: str, *, body: dict | None = None) -> dict:
        """Make a raw API call and return the full response as a dict."""
        resp = await self._http.request(method, path, json=body)
        return {
            "status": resp.status_code,
            "body": resp.json() if resp.headers.get("content-type", "").startswith("application/json") else resp.text,
        }
