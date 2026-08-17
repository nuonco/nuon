import type {
  TBundle,
  TBundleCandidate,
  TBundleUploadStatus,
  TCatalog,
  TCompositePlan,
  THealth,
  TInstallStack,
  TJobLog,
  TJobLogSummary,
  TPlan,
  TReport,
  TRunnerHeartbeat,
  TRun,
  TStackOutputs,
  TStatus,
  TStackCandidate,
  TStepResult,
} from "./types";

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

export const getCatalog = () => request<TCatalog>("/api/catalog");
export const getBundle = () => request<TBundle>("/api/bundle");
export const getBundleUploadStatus = () =>
  request<TBundleUploadStatus>("/api/bundle-candidate/upload-status");
export const getHealth = () => request<THealth>("/api/health");
export const getRunnerHeartbeat = () =>
  request<TRunnerHeartbeat | null>("/api/runner-heartbeat");
export const getStatus = () => request<TStatus>("/api/status");
export const getReport = () => request<TReport>("/api/report");
export const getStackOutputs = () =>
  request<TStackOutputs>("/api/stack-outputs");
export const getInstallStack = () =>
  request<TInstallStack | null>("/api/install-stack");
export const getRuns = () => request<TRun[]>("/api/runs");
export const getRun = (id: string) =>
  request<TRun>(`/api/runs/${encodeURIComponent(id)}`);
export const controlRun = (id: string, action: "retry" | "user-skip" | "cancel") =>
  request(`/api/runs/${encodeURIComponent(id)}/control`, {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ action }),
  });
export const getLogs = () => request<{ jobs: TJobLogSummary[] }>("/api/logs");
export const getJobLog = (id: string) =>
  request<TJobLog>(`/api/logs/${encodeURIComponent(id)}`);
export const jobLogDownloadURL = (id: string) =>
  `/api/logs/${encodeURIComponent(id)}?raw=1`;
export const planDownloadURL = (jobId: string) =>
  `/api/plans/${encodeURIComponent(jobId)}`;
export const getPlan = (jobId: string) =>
  request<TPlan>(planDownloadURL(jobId));
export const stepPlanDownloadURL = (stepId: string) =>
  `/api/step-plans/${encodeURIComponent(stepId)}`;
export const getStepPlan = (stepId: string) =>
  request<TCompositePlan>(stepPlanDownloadURL(stepId));
export const getStepResult = (stepId: string) =>
  request<TStepResult>(`/api/step-results/${encodeURIComponent(stepId)}`);

const csrfHeaders = () => ({
  "Content-Type": "application/json",
  "X-CSRF-Token":
    document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')?.content ??
    "",
});

export const approveBundleCandidate = (bundleDigest: string) =>
  request<{ bundle_digest: string; status: string }>("/api/bundle-candidate/approve", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ bundle_digest: bundleDigest }),
  });

export const clearBundleCandidate = (bundleDigest: string) =>
  request<{ bundle_digest: string; status: string }>("/api/bundle-candidate/clear", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ bundle_digest: bundleDigest }),
  });

export const planBundleCandidateStack = (bundleDigest: string) =>
  request<TStackCandidate>("/api/bundle-candidate/plan-stack", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ bundle_digest: bundleDigest }),
  });

export const uploadBundleCandidate = (
  file: File,
  onProgress: (loaded: number, total: number) => void,
) =>
  new Promise<TBundleCandidate>((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open("POST", "/api/bundle-candidate");
    request.setRequestHeader("Content-Type", "application/zstd");
    request.setRequestHeader("X-Nuon-Bundle-Filename", encodeURIComponent(file.name));
    request.setRequestHeader("X-CSRF-Token", csrfHeaders()["X-CSRF-Token"]);
    request.upload.onprogress = (event) => onProgress(event.loaded, event.lengthComputable ? event.total : file.size);
    request.onerror = () => reject(new Error("Bundle upload failed"));
    request.onload = () => {
      let body: (TBundleCandidate & { error?: string }) | null = null;
      try {
        body = JSON.parse(request.responseText || "null") as TBundleCandidate & { error?: string };
      } catch {
        reject(new Error(`${request.status} ${request.statusText}`));
        return;
      }
      if (request.status < 200 || request.status >= 300) {
        reject(new Error(body?.error ?? `${request.status} ${request.statusText}`));
        return;
      }
      if (!body) {
        reject(new Error("Bundle upload returned an empty response"));
        return;
      }
      resolve(body);
    };
    request.send(file);
  });

export const dispatchRef = (refId: string) => {
  return request<{ dispatch_id: string }>("/api/dispatch", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ ref_id: refId }),
  });
};
