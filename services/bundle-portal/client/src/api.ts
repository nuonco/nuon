import type {
  TConnectedRelease,
  TConnectedReleaseUpdate,
  TConnectedLog,
  TConnectedWorkflow,
  TPortalBranding,
} from "./types";
import type {
  TAppReleaseWithFiles,
  TReleaseFileContent,
  TReleasePackage,
} from "@/types/ctl-api.types";

const request = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init);
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? `${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<T>;
};

const csrfHeaders = () => ({
  "Content-Type": "application/json",
  "X-CSRF-Token":
    document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')
      ?.content ?? "",
});

export const getPortalBranding = () =>
  request<TPortalBranding>("/api/branding");
export const getConnectedReleases = () =>
  request<TConnectedRelease[]>("/api/connected/releases");
export const getConnectedReleaseUpdates = () =>
  request<TConnectedReleaseUpdate[]>("/api/connected/release-updates");
export const getConnectedRelease = (releaseId: string) =>
  request<TAppReleaseWithFiles>(
    `/api/connected/releases/${encodeURIComponent(releaseId)}`,
  );
export const getConnectedReleaseFileContent = (
  releaseId: string,
  path: string,
) =>
  request<TReleaseFileContent>(
    `/api/connected/releases/${encodeURIComponent(releaseId)}/files/content?path=${encodeURIComponent(path)}`,
  );
export const getConnectedReleasePackage = (packageId: string) =>
  request<TReleasePackage>(
    `/api/connected/release-packages/${encodeURIComponent(packageId)}`,
  );
export const getConnectedWorkflows = () =>
  request<TConnectedWorkflow[]>("/api/connected/workflows");
export const getConnectedWorkflow = (workflowId: string) =>
  request<TConnectedWorkflow>(
    `/api/connected/workflows/${encodeURIComponent(workflowId)}`,
  );
export const getConnectedLogStreamLogs = (logStreamId: string) =>
  request<TConnectedLog[]>(
    `/api/connected/log-streams/${encodeURIComponent(logStreamId)}/logs`,
  );
export const retryConnectedWorkflowStep = (
  workflowId: string,
  stepId: string,
) =>
  request<{ workflow_id: string; retryable: boolean }>(
    `/api/connected/workflows/${encodeURIComponent(workflowId)}/steps/${encodeURIComponent(stepId)}/retry`,
    {
      method: "POST",
      headers: csrfHeaders(),
      body: JSON.stringify({}),
    },
  );

const connectedApprovalPath = (
  workflowId: string,
  stepId: string,
  approvalId: string,
) =>
  `/api/connected/workflows/${encodeURIComponent(workflowId)}/steps/${encodeURIComponent(stepId)}/approvals/${encodeURIComponent(approvalId)}`;

export const getConnectedApprovalContents = (
  workflowId: string,
  stepId: string,
  approvalId: string,
) =>
  request<unknown>(
    `${connectedApprovalPath(workflowId, stepId, approvalId)}/contents`,
  );
export const respondToConnectedApproval = (
  workflowId: string,
  stepId: string,
  approvalId: string,
  responseType: "approve" | "deny",
  note = "",
) =>
  request(`${connectedApprovalPath(workflowId, stepId, approvalId)}/response`, {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ response_type: responseType, note }),
  });

export const deployConnectedRelease = (releaseId: string) =>
  request<{ install_app_config_version: unknown; workflow_id: string }>(
    `/api/connected/releases/${encodeURIComponent(releaseId)}/deploy`,
    {
      method: "POST",
      headers: csrfHeaders(),
      body: JSON.stringify({}),
    },
  );
