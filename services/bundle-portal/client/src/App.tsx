import { Fragment, useEffect, useRef, useState, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Link,
  Navigate,
  NavLink,
  Route,
  Routes,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router";
import {
  ArrowLeft,
  ArrowUpRight,
  BookOpen,
  Box,
  CheckCircle2,
  ChevronRight,
  Circle,
  CircleDot,
  Clock,
  Container,
  Copy,
  Cpu,
  Download,
  FileArchive,
  FileCode,
  LayoutDashboard,
  Layers,
  Loader2,
  Package,
  PlayCircle,
  Radar,
  ScrollText,
  Search,
  Server,
  ShieldCheck,
  Terminal,
  UploadCloud,
  WifiOff,
  XCircle,
  Zap,
} from "lucide-react";
import {
  approveBundleCandidate,
  clearBundleCandidate,
  controlRun,
  dispatchRef,
  getBundle,
  getBundleUploadStatus,
  getCatalog,
  getHealth,
  getInstallStack,
  getJobLog,
  getLogs,
  getPlan,
  getReport,
  getRunnerHeartbeat,
  getRun,
  getRuns,
  getStackOutputs,
  getStatus,
  getStepPlan,
  getStepResult,
  jobLogDownloadURL,
  planDownloadURL,
  planBundleCandidateStack,
  stepPlanDownloadURL,
  uploadBundleCandidate,
} from "./api";
import { findPlanResource, PlanResourceDiff, ValueDiff } from "./plan-diff";
import type {
  TBundleActionDefinition,
  TBundleActionStep,
  TBundleRunbookDefinition,
  TBundleContent,
  TBundleChange,
  TBundleHistoryComparison,
  TBundleInfo,
  TCatalogRef,
  TComponentHealth,
  TDrift,
  TJobLogSummary,
  TLogEntry,
  TManifestDiffContent,
  TPlan,
  TResourceDiff,
  TRun,
  TRunStep,
  TStackCandidate,
  TStackChange,
  TStackPropertyChange,
  TStatusStep,
  TStepResult,
} from "./types";

const formatDate = (value?: string) =>
  value
    ? new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "medium",
      }).format(new Date(value))
    : "—";

const formatTime = (value?: string) =>
  value
    ? new Intl.DateTimeFormat(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      }).format(new Date(value))
    : "—";

const statusTheme = (status?: string) => {
  const normalized = status?.toLowerCase() ?? "";
  if (
    ["healthy", "finished", "succeeded", "success", "no drift"].includes(
      normalized,
    )
  )
    return "success";
  if (["failed", "error", "unhealthy", "drifted"].includes(normalized))
    return "error";
  if (["degraded", "warning", "stale"].includes(normalized)) return "warn";
  if (["running", "pending", "queued", "in-progress"].includes(normalized))
    return "info";
  return "";
};

const installStackPhaseLabel = (status: string) => {
  const normalized = status.toUpperCase();
  if (normalized.includes("FAILED") || normalized.includes("ROLLBACK"))
    return "failed";
  if (normalized.endsWith("_IN_PROGRESS")) return "in-progress";
  if (normalized.endsWith("_COMPLETE")) return "finished";
  return "pending";
};

const Badge = ({
  children,
  theme,
  mono,
}: {
  children: ReactNode;
  theme?: string;
  mono?: boolean;
}) => (
  <span
    className={`badge ${theme ?? statusTheme(String(children))} ${mono ? "mono" : ""}`}
  >
    {children}
  </span>
);

const KindIcon = ({ kind }: { kind?: string }) => {
  const k = kind?.toLowerCase() ?? "";
  if (k.includes("installation")) return <Package />;
  if (k.includes("upgrade")) return <Layers />;
  if (k.includes("runbook")) return <BookOpen />;
  if (k.includes("drift")) return <Radar />;
  if (k.includes("cron")) return <Clock />;
  return <Zap />;
};

const State = ({
  error,
  loading,
  empty,
  emptyText,
}: {
  error?: Error | null;
  loading?: boolean;
  empty?: boolean;
  emptyText?: string;
}) => {
  if (loading) return <div className="state">Loading…</div>;
  if (error) return <div className="state error-text">{error.message}</div>;
  if (empty) return <div className="state">{emptyText ?? "No data found"}</div>;
  return null;
};

const PageHeader = ({
  title,
  subtitle,
  actions,
}: {
  title: string;
  subtitle: string;
  actions?: ReactNode;
}) => (
  <div className="page-heading">
    <div>
      <h1>{title}</h1>
      <p>{subtitle}</p>
    </div>
    {actions ?? (
      <span className="polling">
        <i /> refreshing every 5s
      </span>
    )}
  </div>
);

const RUNNER_STALE_MS = 3 * 60 * 1000;

const useRunnerState = () => {
  const heartbeat = useQuery({
    queryKey: ["runner-heartbeat"],
    queryFn: getRunnerHeartbeat,
  });
  const observedAt = heartbeat.data?.observed_at;
  if (!observedAt) return { state: "unknown" as const, observedAt };
  const observedAtMs = Date.parse(observedAt);
  if (!Number.isFinite(observedAtMs))
    return { state: "unknown" as const, observedAt };
  const offline = Date.now() - observedAtMs > RUNNER_STALE_MS;
  return {
    state: offline ? ("offline" as const) : ("active" as const),
    observedAt,
  };
};

const latestRunByRef = (runs: TRun[] | undefined) => {
  const byRef = new Map<string, TRun>();
  for (const run of runs ?? []) {
    const existing = byRef.get(run.ref_id);
    if (
      !existing ||
      Date.parse(run.started_at) > Date.parse(existing.started_at)
    )
      byRef.set(run.ref_id, run);
  }
  return byRef;
};

/* Install progress */

const compositePlanKinds = (plan: Record<string, unknown>) =>
  Object.entries(plan)
    .filter(([key, value]) => key.endsWith("_plan") && value != null)
    .map(([key]) => key.replace(/_plan$/, "").replaceAll("_", " "));

const terraformAction = (actions?: string[]) => {
  if (!actions || actions.length === 0) return "update";
  if (actions.includes("create") && actions.includes("delete"))
    return "replace";
  const action = actions[0];
  return action === "delete" ? "destroy" : action;
};

const TerraformPlanView = ({ plan }: { plan: TPlan }) => {
  const changes = (plan.resource_changes ?? []).filter((change) => {
    const actions = change.change?.actions ?? [];
    return actions.some((action) => action !== "no-op" && action !== "read");
  });
  if (changes.length === 0)
    return (
      <div className="plan-diff plan-diff-empty">
        terraform plan reports no resource changes.
      </div>
    );
  return (
    <ul className="drift-resource-list">
      {changes.map((change) => {
        const action = terraformAction(change.change?.actions);
        return (
          <li key={change.address} className="drift-resource-item">
            <div className="drift-resource-toggle expanded">
              <ChangeBadge action={action} />
              <code>{change.address}</code>
            </div>
            <div className="drift-resource-detail">
              <PlanResourceDiff resource={change} action={action} changedOnly />
            </div>
          </li>
        );
      })}
    </ul>
  );
};

const manifestAction = (diff: TResourceDiff) => {
  const value = diff.op ?? diff.type ?? "";
  const op = typeof value === "string" ? value.toLowerCase() : "";
  if (["create", "install", "apply", "added"].includes(op)) return "create";
  if (["delete", "uninstall", "destroy", "removed"].includes(op))
    return "destroy";
  return "update";
};

const manifestResourceKey = (diff: TResourceDiff, index: number) =>
  `${diff.kind ?? diff.resource ?? "resource"}/${diff.namespace ?? ""}/${diff.name ?? index}`;

// A manifest diff entry carries either structured original→applied values or a
// raw diff payload; fold the structured ones into before/after objects so the
// shared terraform-style attribute diff can render them.
const manifestEntryObjects = (diff: TResourceDiff) => {
  const before: Record<string, unknown> = {};
  const after: Record<string, unknown> = {};
  const payloads: Array<{ line: string; type: string | number }> = [];
  (diff.entries ?? []).forEach((entry, index) => {
    if (entry.payload) {
      payloads.push({ line: entry.payload, type: entry.type });
      return;
    }
    const key = entry.path || `${entry.type}-${index}`;
    if (entry.original !== undefined) before[key] = entry.original;
    if (entry.applied !== undefined) after[key] = entry.applied;
    if (entry.changes && entry.applied === undefined)
      after[key] = entry.changes;
  });
  return { before, after, payloads };
};

const ManifestDiffView = ({ content }: { content: TManifestDiffContent }) => {
  const diffs = content.helm_content_diff ?? content.k8s_content_diff ?? [];
  if (diffs.length === 0) {
    return (
      <div className="plan-diff plan-diff-empty">
        {content.plan && content.plan !== "no changes"
          ? content.plan
          : "No resource changes — the manifest matches the cluster."}
      </div>
    );
  }
  return (
    <ul className="drift-resource-list">
      {diffs.map((diff, index) => {
        const key = manifestResourceKey(diff, index);
        const action = manifestAction(diff);
        const label = [
          diff.kind,
          diff.namespace && diff.name
            ? `${diff.namespace}/${diff.name}`
            : diff.name,
        ]
          .filter(Boolean)
          .join(" ");
        return (
          <li key={key} className="drift-resource-item">
            <div className="drift-resource-toggle expanded">
              <ChangeBadge action={action} />
              <code>{label || key}</code>
              {diff.error && (
                <Badge theme="error" mono>
                  error
                </Badge>
              )}
            </div>
            <div className="drift-resource-detail">
              {diff.error && (
                <div className="plan-diff plan-diff-empty">{diff.error}</div>
              )}
              <ManifestResourceDetail diff={diff} action={action} />
            </div>
          </li>
        );
      })}
    </ul>
  );
};

const ManifestResourceDetail = ({
  diff,
  action,
}: {
  diff: TResourceDiff;
  action: string;
}) => {
  const { before, after, payloads } = manifestEntryObjects(diff);
  const hasStructured =
    Object.keys(before).length > 0 || Object.keys(after).length > 0;
  return (
    <>
      {hasStructured && (
        <PlanResourceDiff
          resource={{
            address: manifestResourceKey(diff, 0),
            change: { before, after },
          }}
          action={action}
          changedOnly
        />
      )}
      {payloads.length > 0 && <ManifestLineDiff payloads={payloads} />}
      {!hasStructured && payloads.length === 0 && (
        <div className="plan-diff plan-diff-empty">
          No detailed changes recorded for this resource.
        </div>
      )}
    </>
  );
};

const ManifestLineDiff = ({ payloads, changedOnly = true }: { payloads: Array<{ line: string; type: string | number }>; changedOnly?: boolean }) => {
  const [showContext, setShowContext] = useState(false);
  const firstChangedLine = useRef<HTMLSpanElement>(null);
  const hiddenContext = changedOnly ? payloads.filter(
    (payload) => payload.type !== 1 && payload.type !== 2 && payload.type !== "removed" && payload.type !== "added",
  ).length : 0;
  const visible = showContext || !changedOnly
    ? payloads
    : payloads.filter(
        (payload) => payload.type === 1 || payload.type === 2 || payload.type === "removed" || payload.type === "added",
      );
  const firstChangedIndex = visible.findIndex(
    (payload) => payload.type === 1 || payload.type === 2 || payload.type === "removed" || payload.type === "added",
  );
  return (
    <div className="manifest-line-diff-wrap">
      {hiddenContext > 0 && (
        <button type="button" className="plan-diff-context" onClick={() => {
          const next = !showContext;
          setShowContext(next);
          if (next) requestAnimationFrame(() => firstChangedLine.current?.scrollIntoView({ block: "center" }));
        }}>
          {showContext ? "Hide unchanged lines" : `Show ${hiddenContext} unchanged ${hiddenContext === 1 ? "line" : "lines"}`}
        </button>
      )}
      <pre className="manifest-line-diff">
        {visible.map((payload, index) => {
          const removed = payload.type === 1 || payload.type === "removed";
          const added = payload.type === 2 || payload.type === "added";
          return (
            <span ref={index === firstChangedIndex ? firstChangedLine : undefined} key={index} className={removed ? "removed" : added ? "added" : "context"}>
              <b>{removed ? "−" : added ? "+" : " "}</b>{payload.line}{"\n"}
            </span>
          );
        })}
      </pre>
    </div>
  );
};

const stepResultLabel = (kind: TStepResult["kind"]) =>
  kind === "terraform"
    ? "terraform plan computed by the airgapped runner"
    : kind === "helm"
      ? "helm diff computed by the airgapped runner"
      : kind === "kubernetes_manifest"
        ? "kubernetes manifest diff computed by the airgapped runner"
        : "execution result recorded by the airgapped runner";

const StepPlanPreview = ({ step }: { step: TStatusStep & { job_id?: string } }) => {
	const resultID = step.job_id ?? step.id;
  const finished = step.status === "finished";
  const result = useQuery({
    queryKey: ["step-result", resultID],
    queryFn: () => getStepResult(resultID),
    retry: false,
    refetchInterval: (query) =>
      query.state.data || (finished && query.state.error) ? false : 5000,
  });
  const plan = useQuery({
    queryKey: ["step-plan", resultID],
    queryFn: () => getStepPlan(resultID),
    retry: false,
    refetchInterval: (query) => (query.state.data ? false : 5000),
  });
  const [showRaw, setShowRaw] = useState(false);
  const waiting = step.status === "available";

  const resultContent =
    result.data && typeof result.data.content === "object"
      ? result.data.content
      : null;
  const kinds = compositePlanKinds(plan.data ?? {});

  const rawPlanToggle = plan.data ? (
    <div style={{ paddingTop: 8 }}>
      <button
        type="button"
        className="link"
        style={{
          fontSize: 12,
          background: "none",
          border: "none",
          padding: 0,
          cursor: "pointer",
        }}
        onClick={() => setShowRaw((current) => !current)}
      >
        {showRaw ? "hide" : "show"} raw execution plan
      </button>
      {showRaw && (
        <pre
          style={{
            margin: "8px 0 0",
            padding: 12,
            fontSize: 12,
            maxHeight: 360,
            overflow: "auto",
            background: "var(--surface-sunken, rgba(0,0,0,0.2))",
            borderRadius: 6,
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
          }}
        >
          {JSON.stringify(plan.data, null, 2)}
        </pre>
      )}
    </div>
  ) : null;

  if (resultContent && result.data) {
    const kind = result.data.kind;
    return (
      <div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            padding: "8px 0",
          }}
        >
          <Badge mono>
            {kind === "kubernetes_manifest" ? "kubernetes manifest" : kind}
          </Badge>
          {!result.data.success && <Badge theme="error">failed</Badge>}
          <span style={{ fontSize: 12, color: "var(--muted)" }}>
            {stepResultLabel(kind)}
          </span>
          <a
            className="link"
            style={{ marginLeft: "auto", fontSize: 12 }}
            href={`/api/step-results/${encodeURIComponent(resultID)}`}
            download={`${step.id}-result.json`}
          >
            <Download style={{ width: 13, height: 13 }} /> result.json
          </a>
        </div>
        {kind === "terraform" ? (
          <TerraformPlanView plan={resultContent as TPlan} />
        ) : kind === "helm" || kind === "kubernetes_manifest" ? (
          <ManifestDiffView content={resultContent as TManifestDiffContent} />
        ) : (
          <pre
            style={{
              margin: 0,
              padding: 12,
              fontSize: 12,
              maxHeight: 360,
              overflow: "auto",
              background: "var(--surface-sunken, rgba(0,0,0,0.2))",
              borderRadius: 6,
              whiteSpace: "pre-wrap",
              wordBreak: "break-word",
            }}
          >
            {JSON.stringify(resultContent, null, 2)}
          </pre>
        )}
        {rawPlanToggle}
      </div>
    );
  }

  if (plan.isLoading || result.isLoading)
    return <div className="plan-diff plan-diff-empty">Loading plan…</div>;
  if (plan.isError) {
    return (
      <div className="plan-diff plan-diff-empty">
        {waiting
          ? "Plan not rendered yet — the runner renders each step’s late-bound plan once earlier steps publish the outputs it needs."
          : "No rendered plan recorded for this step."}
      </div>
    );
  }
  return (
    <div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          padding: "8px 0",
        }}
      >
        {kinds.map((kind) => (
          <Badge key={kind} mono>
            {kind}
          </Badge>
        ))}
        <span style={{ fontSize: 12, color: "var(--muted)" }}>
          {finished
            ? "late-bound plan rendered by the airgapped runner"
            : "late-bound plan rendered by the airgapped runner — the real terraform plan / diff appears once the step’s plan phase completes"}
        </span>
        <a
          className="link"
          style={{ marginLeft: "auto", fontSize: 12 }}
          href={stepPlanDownloadURL(step.id)}
          download={`${step.id}-plan.json`}
        >
          <Download style={{ width: 13, height: 13 }} /> plan.json
        </a>
      </div>
      <pre
        style={{
          margin: 0,
          padding: 12,
          fontSize: 12,
          maxHeight: 360,
          overflow: "auto",
          background: "var(--surface-sunken, rgba(0,0,0,0.2))",
          borderRadius: 6,
          whiteSpace: "pre-wrap",
          wordBreak: "break-word",
        }}
      >
        {JSON.stringify(plan.data, null, 2)}
      </pre>
    </div>
  );
};

const CandidatePlanResults = ({
  title,
  description,
  changes,
  steps,
}: {
  title: string;
  description: string;
  changes: TBundleChange[];
  steps: TRunStep[];
}) => (
  <section className="candidate-plan-results">
    <div className="candidate-plan-results-heading">
      <div><h3>{title}</h3><p>{description}</p></div>
      <Badge>{changes.length}</Badge>
    </div>
    {changes.length === 0 ? (
      <div className="notice">No changed resources require a plan.</div>
    ) : (
      <div className="candidate-plan-result-list">
        {changes.map((change) => {
          const step = change.plan_step_id
            ? steps.find((candidate) => candidate.id === change.plan_step_id)
            : null;
          return (
            <article className="candidate-plan-result" key={`${change.kind}-${change.name}`}>
              <header className="candidate-plan-result-header">
                <span><ContentKindIcon kind={change.kind} /><strong>{change.name}</strong><small>{change.detail || change.kind.replaceAll("-", " ")}</small></span>
                <span className="candidate-plan-result-state">
                  <Badge theme={step?.status === "failed" ? "error" : step?.status === "finished" ? "success" : "info"} mono>{step?.status ?? "not planned"}</Badge>
                </span>
              </header>
              <div className="candidate-plan-result-body">
                {step?.job_id
                  ? <StepPlanPreview step={step} />
                  : step
                    ? <div className="notice">Waiting for the candidate runner to start this plan.</div>
                    : <div className="notice">No candidate deployment plan has run for this change.</div>}
              </div>
            </article>
          );
        })}
      </div>
    )}
  </section>
);

const InstallProgress = () => {
  const status = useQuery({
    queryKey: ["status"],
    queryFn: getStatus,
    retry: false,
    refetchInterval: (query) =>
      query.state.data?.status === "finished" ||
      query.state.data?.status === "failed"
        ? false
        : 5000,
  });
  const installStack = useQuery({
    queryKey: ["install-stack"],
    queryFn: getInstallStack,
    retry: false,
    refetchInterval: (query) =>
      query.state.data?.phase === "finished" ||
      query.state.data?.phase === "failed"
        ? false
        : 5000,
  });
  const [expanded, setExpanded] = useState<string | null>(null);

  if (
    (status.isLoading || status.isError || !status.data) &&
    (installStack.isLoading || installStack.isError || !installStack.data)
  )
    return null;

  const steps = status.data?.steps ?? [];
  const stack = installStack.data;
  const stackFinished = stack?.phase === "finished";
  const finished =
    steps.filter((step) => step.status === "finished").length +
    (stackFinished ? 1 : 0);
  const total = steps.length + (stack ? 1 : 0);
  const overallStatus = status.data?.status ?? stack?.phase ?? "pending";

  return (
    <section className="card" style={{ marginBottom: 20 }}>
      <div className="card-header">
        <h3>
          <Terminal /> Install progress
        </h3>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span style={{ fontSize: 12, color: "var(--muted)" }}>
            {finished}/{total} steps
          </span>
          <Badge>{overallStatus}</Badge>
        </div>
      </div>
      {status.data?.failed_step && (
        <div className="state error-text">
          Deployment failed at step {status.data.failed_step} — expand it below
          for the error and logs.
        </div>
      )}
      {stack?.phase === "failed" && (
        <div className="state error-text">
          Install stack failed: {stack.status_reason ?? stack.status}
        </div>
      )}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th style={{ width: "40%" }}>Step</th>
              <th>Status</th>
              <th>Started</th>
              <th>Duration</th>
              <th>Logs</th>
            </tr>
          </thead>
          <tbody>
            {stack && (
              <Fragment>
                <tr
                  className={`selectable ${expanded === "__install-stack" ? "selected" : ""}`}
                  onClick={() =>
                    setExpanded(
                      expanded === "__install-stack" ? null : "__install-stack",
                    )
                  }
                >
                  <td>
                    <span className="strong">
                      {expanded === "__install-stack" ? "▾ " : "▸ "}1. Provision
                      install stack
                      {stack.phase === "in-progress" && (
                        <Loader2
                          className="spin"
                          style={{
                            width: 13,
                            height: 13,
                            marginLeft: 6,
                            verticalAlign: -2,
                          }}
                        />
                      )}
                    </span>
                    <code className="subtle">{stack.name}</code>
                  </td>
                  <td>
                    <Badge theme={statusTheme(stack.phase)}>
                      {stack.status}
                    </Badge>
                    {stack.status_reason && (
                      <div
                        style={{
                          fontSize: 12,
                          color: "var(--error-text)",
                          marginTop: 4,
                        }}
                      >
                        {stack.status_reason}
                      </div>
                    )}
                  </td>
                  <td style={{ fontSize: 12, color: "var(--muted)" }}>
                    {formatTime(stack.started_at)}
                  </td>
                  <td style={{ fontSize: 12, color: "var(--muted)" }}>
                    {stack.phase === "in-progress"
                      ? "running…"
                      : formatDuration(stack.started_at, stack.updated_at)}
                  </td>
                  <td>—</td>
                </tr>
                {expanded === "__install-stack" && (
                  <tr>
                    <td colSpan={5}>
                      {stack.events.length === 0 ? (
                        <div className="state">
                          Waiting for CloudFormation stack creation.
                        </div>
                      ) : (
                        <div style={{ display: "grid", gap: 8 }}>
                          {stack.events.slice(-50).map((event) => (
                            <div
                              key={event.id}
                              style={{
                                display: "grid",
                                gridTemplateColumns:
                                  "145px minmax(180px, 1fr) 180px",
                                gap: 12,
                                fontSize: 12,
                              }}
                            >
                              <span style={{ color: "var(--muted)" }}>
                                {formatTime(event.timestamp)}
                              </span>
                              <span>
                                <strong>
                                  {event.logical_resource_id ?? stack.name}
                                </strong>
                                {event.status_reason && (
                                  <span className="error-text">
                                    {" "}
                                    — {event.status_reason}
                                  </span>
                                )}
                              </span>
                              <Badge
                                theme={statusTheme(
                                  installStackPhaseLabel(event.status),
                                )}
                              >
                                {event.status}
                              </Badge>
                            </div>
                          ))}
                        </div>
                      )}
                    </td>
                  </tr>
                )}
              </Fragment>
            )}
            {steps.map((step, index) => {
              const isExpanded = expanded === step.id;
              const running = step.status === "in-progress";
              return (
                <Fragment key={step.id}>
                  <tr
                    className={`selectable ${isExpanded ? "selected" : ""}`}
                    onClick={() => setExpanded(isExpanded ? null : step.id)}
                  >
                    <td>
                      <span className="strong">
                        {isExpanded ? "▾ " : "▸ "}
                        {index + 1 + (stack ? 1 : 0)}. {step.name}
                        {running && (
                          <Loader2
                            className="spin"
                            style={{
                              width: 13,
                              height: 13,
                              marginLeft: 6,
                              verticalAlign: -2,
                            }}
                          />
                        )}
                      </span>
                      <code className="subtle">{step.id}</code>
                    </td>
                    <td>
                      <Badge>{step.status}</Badge>
                      {step.error && (
                        <div
                          style={{
                            fontSize: 12,
                            color: "var(--error-text)",
                            marginTop: 4,
                          }}
                        >
                          {step.error}
                        </div>
                      )}
                    </td>
                    <td style={{ fontSize: 12, color: "var(--muted)" }}>
                      {formatTime(step.started_at)}
                    </td>
                    <td style={{ fontSize: 12, color: "var(--muted)" }}>
                      {running
                        ? "running…"
                        : formatDuration(step.started_at, step.finished_at)}
                    </td>
                    <td onClick={(event) => event.stopPropagation()}>
                      <Link
                        className="link"
                        to={`/logs?job=${encodeURIComponent(step.id)}`}
                      >
                        logs →
                      </Link>
                    </td>
                  </tr>
                  {isExpanded && (
                    <tr>
                      <td colSpan={5}>
                        <StepPlanPreview step={step} />
                      </td>
                    </tr>
                  )}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
};

/* Dashboard */

const Dashboard = () => {
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const health = useQuery({ queryKey: ["health"], queryFn: getHealth });
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const runner = useRunnerState();

  const components = health.data?.latest?.components ?? [];
  const healthy = components.filter(
    (c) => c.health?.toLowerCase() === "healthy",
  ).length;
  const refs = catalog.data?.refs ?? [];
  const countBy = (match: (ref: TCatalogRef) => boolean) =>
    refs.filter(match).length;
  const sorted = [...(runs.data ?? [])].sort(
    (a, b) => Date.parse(b.started_at) - Date.parse(a.started_at),
  );
  const lastRun = sorted[0];
  const digest = catalog.data?.bundle_digest?.replace("sha256:", "");

  return (
    <main>
      <PageHeader
        title="Dashboard"
        subtitle="Bundle, component health, and recent activity for this environment."
      />

      <div className="tiles">
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Components</span>
            {components.length > 0 && healthy === components.length ? (
              <CheckCircle2 className="good" />
            ) : (
              <XCircle className={components.length === 0 ? "" : "bad"} />
            )}
          </div>
          <div className="tile-value">
            {components.length === 0
              ? "—"
              : `${healthy}/${components.length} healthy`}
          </div>
          <div className="tile-note">reported by airgapped runner</div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Runner</span>
            <CircleDot
              className={
                runner.state === "active"
                  ? "good"
                  : runner.state === "offline"
                    ? "warn"
                    : ""
              }
            />
          </div>
          <div className="tile-value">{runner.state}</div>
          <div className="tile-note">
            last observed {formatTime(runner.observedAt)}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Bundle</span>
            <ShieldCheck className="good" />
          </div>
          <div className="tile-value">
            {digest ? digest.slice(0, 12) + "…" : "—"}
          </div>
          <div className="tile-note">checksum-verified envelope</div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Last run</span>
            {lastRun ? <Badge>{lastRun.status}</Badge> : <Clock />}
          </div>
          <div className="tile-value">
            {lastRun ? lastRun.ref_name || lastRun.ref_id : "none yet"}
          </div>
          <div className="tile-note">
            {lastRun
              ? formatTime(lastRun.started_at)
              : "dispatch from Runbooks"}
          </div>
        </div>
      </div>

      <div className="grid-2-1">
        <div className="stack">
          <section className="card">
            <div className="card-header">
              <h3>Active bundle</h3>
              <Badge theme="success" mono>
                verified
              </Badge>
            </div>
            <div className="card-body">
              <State error={catalog.error} loading={catalog.isLoading} />
              {catalog.data && (
                <dl className="kv">
                  <div>
                    <dt>deployment</dt>
                    <dd>{catalog.data.deployment_id}</dd>
                  </div>
                  <div>
                    <dt>sha256</dt>
                    <dd title={catalog.data.bundle_digest}>
                      {digest?.slice(0, 20)}…
                    </dd>
                  </div>
                  <div>
                    <dt>generated</dt>
                    <dd>{formatTime(catalog.data.generated_at)}</dd>
                  </div>
                  <hr />
                  <div>
                    <dt>components</dt>
                    <dd>{components.length || "—"}</dd>
                  </div>
                  <div>
                    <dt>actions</dt>
                    <dd>
                      {countBy((r) => r.kind.toLowerCase().includes("action"))}
                    </dd>
                  </div>
                  <div>
                    <dt>runbooks</dt>
                    <dd>
                      {countBy((r) => r.kind.toLowerCase().includes("runbook"))}
                    </dd>
                  </div>
                  <div>
                    <dt>drift checks</dt>
                    <dd>
                      {countBy((r) => r.kind.toLowerCase().includes("drift"))}
                    </dd>
                  </div>
                </dl>
              )}
            </div>
          </section>

          <div className="quick-actions" style={{ gridTemplateColumns: "1fr" }}>
            <NavLink className="quick-action" to="/runbooks">
              <span className="qa-icon">
                <PlayCircle />
              </span>
              <span className="qa-text">
                <div className="qa-title">Run a runbook</div>
                <div className="qa-hint">{refs.length} refs available</div>
              </span>
              <ArrowUpRight />
            </NavLink>
            <NavLink className="quick-action" to="/runs">
              <span className="qa-icon">
                <ScrollText />
              </span>
              <span className="qa-text">
                <div className="qa-title">Inspect runs</div>
                <div className="qa-hint">stages, drift verdicts, job IDs</div>
              </span>
              <ArrowUpRight />
            </NavLink>
          </div>
        </div>

        <div className="stack">
          <section className="card">
            <div className="card-header">
              <h3>Component health</h3>
            </div>
            <State
              error={health.error}
              loading={health.isLoading}
              empty={!health.isLoading && components.length === 0}
              emptyText="No health reports yet"
            />
            {components.length > 0 && (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>Health</th>
                    </tr>
                  </thead>
                  <tbody>
                    {components.map((component) => (
                      <tr
                        key={
                          component.install_component_id ??
                          component.component_id ??
                          component.component_name
                        }
                      >
                        <td>
                          <Link
                            className="strong link"
                            to={componentPath(component)}
                          >
                            {componentKey(component) || "Unnamed component"}
                          </Link>
                        </td>
                        <td>
                          <code>{component.component_type ?? "—"}</code>
                        </td>
                        <td>
                          <Badge>{component.health || "unknown"}</Badge>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="card">
            <div className="card-header">
              <h3>Recent activity</h3>
              <NavLink className="link" to="/runs">
                View all runs →
              </NavLink>
            </div>
            <State
              error={runs.error}
              loading={runs.isLoading}
              empty={!runs.isLoading && sorted.length === 0}
              emptyText="No runs recorded yet"
            />
            <div className="activity">
              {sorted.slice(0, 5).map((run) => (
                <div className="activity-item" key={run.run_id}>
                  <span className="activity-time">
                    {formatTime(run.started_at)}
                  </span>
                  <span
                    className={`status-dot ${statusTheme(run.status) || "neutral"}`}
                  />
                  <span className="activity-body">
                    {run.ref_name || run.ref_id} · {run.ref_kind}
                    <code>{run.run_id}</code>
                  </span>
                  <Badge>{run.status}</Badge>
                </div>
              ))}
            </div>
          </section>
        </div>
      </div>
    </main>
  );
};

/* Components */

const componentKey = (component: TComponentHealth) =>
  component.component_name ??
  component.component_id ??
  component.install_component_id ??
  "";

const componentPath = (component: TComponentHealth) =>
  `/components/${encodeURIComponent(componentKey(component))}`;

const resourceCounts = (component: TComponentHealth) => {
  const resources = component.resources ?? [];
  const healthy = resources.filter(
    (r) => r.health?.toLowerCase() === "healthy",
  ).length;
  return { healthy, total: resources.length };
};

const Components = () => {
  const health = useQuery({ queryKey: ["health"], queryFn: getHealth });
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const components = health.data?.latest?.components ?? [];
  const refs = catalog.data?.refs ?? [];
  const lastRuns = latestRunByRef(runs.data);

  const driftOf = (key: string) => {
    const driftRefs = refs.filter(
      (ref) =>
        ref.component === key && ref.kind.toLowerCase().includes("drift"),
    );
    if (driftRefs.length === 0) return null;
    let latest: { run: TRun; drift: TDrift } | undefined;
    for (const ref of driftRefs) {
      const run = lastRuns.get(ref.id);
      const drift = driftStepOf(run)?.drift;
      if (
        run &&
        drift &&
        (!latest ||
          Date.parse(run.started_at) > Date.parse(latest.run.started_at))
      ) {
        latest = { run, drift };
      }
    }
    return latest ?? "unchecked";
  };

  return (
    <main>
      <PageHeader
        title="Components"
        subtitle="Deployed components, their live health, and the day-2 refs that operate them."
      />
      <section className="card">
        <State
          error={health.error}
          loading={health.isLoading}
          empty={!health.isLoading && !health.error && components.length === 0}
          emptyText="No health reports yet — the airgapped runner publishes them on its health interval"
        />
        {components.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Component</th>
                  <th>Type</th>
                  <th>Health</th>
                  <th>Drift</th>
                  <th>Resources</th>
                  <th>Day-2 refs</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {components.map((component) => {
                  const key = componentKey(component);
                  const counts = resourceCounts(component);
                  const related = refs.filter((ref) => ref.component === key);
                  const drift = driftOf(key);
                  return (
                    <tr key={component.install_component_id ?? key}>
                      <td>
                        <Link
                          className="name-cell link"
                          to={componentPath(component)}
                        >
                          <Box />
                          <span className="strong">
                            {key || "Unnamed component"}
                          </span>
                        </Link>
                      </td>
                      <td>
                        <code>{component.component_type ?? "—"}</code>
                      </td>
                      <td>
                        <Badge>{component.health || "unknown"}</Badge>
                      </td>
                      <td>
                        {drift === null ? (
                          <span style={{ color: "var(--faint)" }}>—</span>
                        ) : drift === "unchecked" ? (
                          <Badge mono>unchecked</Badge>
                        ) : (
                          <Link
                            to={`/runs?run=${encodeURIComponent(drift.run.run_id)}`}
                            title={`checked ${formatTime(drift.run.started_at)}`}
                          >
                            <Badge
                              theme={drift.drift.drifted ? "error" : "success"}
                              mono
                            >
                              {drift.drift.drifted ? "drifted" : "no drift"}
                            </Badge>
                          </Link>
                        )}
                      </td>
                      <td
                        style={{ fontFamily: "var(--font-mono)", fontSize: 12 }}
                      >
                        {counts.total === 0
                          ? "—"
                          : `${counts.healthy}/${counts.total} healthy`}
                      </td>
                      <td>
                        {related.length === 0 ? (
                          <span style={{ color: "var(--faint)" }}>—</span>
                        ) : (
                          <span className="chips">
                            {related.map((ref) => (
                              <Badge key={ref.id} mono>
                                {ref.kind}
                              </Badge>
                            ))}
                          </span>
                        )}
                      </td>
                      <td style={{ textAlign: "right" }}>
                        <Link className="link" to={componentPath(component)}>
                          Details →
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
};

const ComponentDetail = () => {
  const { name = "" } = useParams();
  const health = useQuery({ queryKey: ["health"], queryFn: getHealth });
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const bundle = useQuery({ queryKey: ["bundle"], queryFn: getBundle });
  const status = useQuery({ queryKey: ["status"], queryFn: getStatus });

  const components = health.data?.latest?.components ?? [];
  const component = components.find((c) => componentKey(c) === name);
  const observedAt = health.data?.latest?.observed_at;

  const transitions = (health.data?.transitions ?? [])
    .filter(
      (t) =>
        t.component_name === name ||
        (component?.component_id && t.component_id === component.component_id),
    )
    .sort((a, b) => Date.parse(b.observed_at) - Date.parse(a.observed_at));

  const refs = (catalog.data?.refs ?? []).filter(
    (ref) => ref.component === name,
  );
  const refIDs = new Set(refs.map((ref) => ref.id));
  const componentRuns = (runs.data ?? [])
    .filter((run) => refIDs.has(run.ref_id))
    .sort((a, b) => Date.parse(b.started_at) - Date.parse(a.started_at));

  const inventory = bundle.data?.active?.contents?.find(
    (item) => item.kind === "component" && item.name === name,
  );
  const installSteps = (status.data?.steps ?? []).filter((step) =>
    step.name.toLowerCase().includes(name.toLowerCase()),
  );

  const resources = component?.resources ?? [];
  const counts = component
    ? resourceCounts(component)
    : { healthy: 0, total: 0 };

  if (!health.isLoading && !component) {
    return (
      <main>
        <Link className="link" to="/components">
          <ArrowLeft style={{ width: 14, height: 14 }} /> All components
        </Link>
        <div className="state" style={{ marginTop: 20 }}>
          No component named “{name}” in the latest health report.
        </div>
      </main>
    );
  }

  return (
    <main>
      <div style={{ marginBottom: 12 }}>
        <Link className="link" to="/components">
          <ArrowLeft style={{ width: 14, height: 14 }} /> All components
        </Link>
      </div>
      <PageHeader
        title={name}
        subtitle={`${component?.component_type ?? "component"} · reported by the airgapped runner`}
      />

      <div className="tiles">
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Health</span>
            {component?.health?.toLowerCase() === "healthy" ? (
              <CheckCircle2 className="good" />
            ) : (
              <XCircle className="bad" />
            )}
          </div>
          <div className="tile-value">{component?.health ?? "—"}</div>
          <div className="tile-note">
            last observed {formatTime(observedAt)}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Resources</span>
            <Layers />
          </div>
          <div className="tile-value">
            {counts.total === 0 ? "—" : `${counts.healthy}/${counts.total}`}
          </div>
          <div className="tile-note">
            {component?.truncated
              ? "healthy (list truncated)"
              : "healthy resources"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Day-2 refs</span>
            <Zap />
          </div>
          <div className="tile-value">{refs.length}</div>
          <div className="tile-note">
            {refs.map((ref) => ref.kind).join(" · ") ||
              "none reference this component"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Packaged</span>
            <FileArchive />
          </div>
          <div className="tile-value">
            {inventory ? formatBytes(inventory.size) : "—"}
          </div>
          <div className="tile-note">
            {inventory?.digest
              ? `sha256:${inventory.digest.replace("sha256:", "").slice(0, 12)}…`
              : "not in bundle inventory"}
          </div>
        </div>
      </div>

      <div className="grid-2-1">
        <div className="stack">
          <section className="card">
            <div className="card-header">
              <h3>
                <Server /> Resources
              </h3>
              {component?.truncated && (
                <Badge theme="warn" mono>
                  truncated
                </Badge>
              )}
            </div>
            <State
              loading={health.isLoading}
              empty={!health.isLoading && resources.length === 0}
              emptyText="The health report carries no per-resource detail for this component"
            />
            {resources.length > 0 && (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Resource</th>
                      <th>Namespace</th>
                      <th>Health</th>
                      <th>Message</th>
                    </tr>
                  </thead>
                  <tbody>
                    {resources.map((resource, index) => (
                      <tr key={`${resource.kind}-${resource.name}-${index}`}>
                        <td>
                          <span className="strong">{resource.name ?? "—"}</span>
                          <code className="subtle">
                            {[resource.api_group, resource.kind]
                              .filter(Boolean)
                              .join("/") ||
                              resource.provider ||
                              "—"}
                          </code>
                        </td>
                        <td>
                          <code>{resource.namespace ?? "—"}</code>
                        </td>
                        <td>
                          <Badge>{resource.health ?? "unknown"}</Badge>
                        </td>
                        <td
                          style={{
                            fontSize: 12,
                            color: "var(--muted)",
                            maxWidth: 320,
                          }}
                        >
                          {resource.message ?? "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="card">
            <div className="card-header">
              <h3>
                <ScrollText /> Day-2 activity
              </h3>
              <NavLink className="link" to="/runs">
                All runs →
              </NavLink>
            </div>
            <State
              loading={runs.isLoading}
              empty={!runs.isLoading && componentRuns.length === 0}
              emptyText="No day-2 runs reference this component yet"
            />
            <div className="activity">
              {componentRuns.slice(0, 8).map((run) => (
                <div className="activity-item" key={run.run_id}>
                  <span className="activity-time">
                    {formatTime(run.started_at)}
                  </span>
                  <span
                    className={`status-dot ${statusTheme(run.status) || "neutral"}`}
                  />
                  <span className="activity-body">
                    {run.ref_name || run.ref_id} · {run.ref_kind}
                    <code>
                      {(run.steps ?? [])
                        .filter((step) => step.job_id)
                        .map((step, index) => (
                          <span key={step.id}>
                            {index > 0 && " · "}
                            <Link
                              className="link"
                              to={`/logs?job=${encodeURIComponent(step.job_id!)}`}
                            >
                              {step.job_id}
                            </Link>
                          </span>
                        ))}
                      {(run.steps ?? []).every((step) => !step.job_id) &&
                        run.run_id}
                    </code>
                  </span>
                  <Badge>{run.status}</Badge>
                </div>
              ))}
            </div>
          </section>
        </div>

        <div className="stack">
          <section className="card">
            <div className="card-header">
              <h3>
                <Clock /> Health history
              </h3>
            </div>
            <State
              loading={health.isLoading}
              empty={!health.isLoading && transitions.length === 0}
              emptyText="No health transitions recorded — steady since first report"
            />
            <div className="activity">
              {transitions.slice(0, 10).map((transition, index) => (
                <div
                  className="activity-item"
                  key={`${transition.observed_at}-${index}`}
                >
                  <span className="activity-time">
                    {formatTime(transition.observed_at)}
                  </span>
                  <span
                    className={`status-dot ${statusTheme(transition.to) || "neutral"}`}
                  />
                  <span className="activity-body">
                    <Badge>{transition.from || "unknown"}</Badge> →{" "}
                    <Badge>{transition.to}</Badge>
                  </span>
                </div>
              ))}
            </div>
          </section>

          <section className="card">
            <div className="card-header">
              <h3>
                <Zap /> Day-2 refs
              </h3>
              <NavLink className="link" to="/runbooks">
                Dispatch →
              </NavLink>
            </div>
            <State
              loading={catalog.isLoading}
              empty={!catalog.isLoading && refs.length === 0}
              emptyText="No actions, drift checks, or runbooks reference this component"
            />
            <div className="activity">
              {refs.map((ref) => (
                <div className="activity-item" key={ref.id}>
                  <KindIcon kind={ref.kind} />
                  <span className="activity-body">
                    {ref.name}
                    <code>{ref.id}</code>
                  </span>
                  <Badge mono>{ref.kind}</Badge>
                </div>
              ))}
            </div>
          </section>

          <section className="card">
            <div className="card-header">
              <h3>
                <Terminal /> Install jobs
              </h3>
            </div>
            <State
              loading={status.isLoading}
              empty={!status.isLoading && installSteps.length === 0}
              emptyText="No install steps mention this component"
            />
            <div className="activity">
              {installSteps.map((step) => (
                <div className="activity-item" key={step.id}>
                  <span className="activity-time">
                    {formatTime(step.started_at)}
                  </span>
                  <span
                    className={`status-dot ${statusTheme(step.status) || "neutral"}`}
                  />
                  <span className="activity-body">
                    {step.name}
                    <code>
                      <Link
                        className="link"
                        to={`/logs?job=${encodeURIComponent(step.id)}`}
                      >
                        {step.id}
                      </Link>
                    </code>
                  </span>
                  <Badge>{step.status}</Badge>
                </div>
              ))}
            </div>
          </section>
        </div>
      </div>
    </main>
  );
};

/* Drift resource verdicts */

const changeSymbol = (action: string) =>
  action === "create"
    ? "+"
    : action === "update"
      ? "~"
      : action === "destroy"
        ? "−"
        : action === "replace"
          ? "±"
          : "=";

const ChangeBadge = ({ action }: { action: string }) => (
  <span className={`change-badge ${action}`} title={action}>
    {changeSymbol(action)}
  </span>
);

const driftStepOf = (run?: TRun) => run?.steps?.find((step) => step.drift);

const DriftResources = ({
  drift,
  jobID,
}: {
  drift: TDrift;
  jobID?: string;
}) => {
  const resources = drift.resources ?? [];
  const [expandedAddress, setExpandedAddress] = useState<string | null>(null);
  const planQuery = useQuery({
    queryKey: ["plan", jobID],
    queryFn: () => getPlan(jobID!),
    enabled: !!jobID && expandedAddress !== null,
    staleTime: Infinity,
    retry: false,
  });
  if (resources.length === 0) return null;

  const toggle = (address: string) =>
    setExpandedAddress((current) => (current === address ? null : address));

  return (
    <div className="drift-resources">
      <ul className="drift-resource-list">
        {resources.map((resource) => {
          const expanded = expandedAddress === resource.address;
          const planResource = planQuery.data
            ? findPlanResource(planQuery.data, resource.address)
            : undefined;
          return (
            <li key={resource.address} className="drift-resource-item">
              <button
                type="button"
                className={`drift-resource-toggle${expanded ? " expanded" : ""}`}
                onClick={() => toggle(resource.address)}
                disabled={!jobID}
                title={jobID ? "Show attribute diff" : "Full plan unavailable"}
              >
                <ChevronRight className="drift-resource-chevron" />
                <ChangeBadge action={resource.action} />
                <code>{resource.address}</code>
                {resource.drifted && (
                  <Badge theme="warn" mono>
                    out-of-band drift
                  </Badge>
                )}
              </button>
              {expanded && (
                <div className="drift-resource-detail">
                  {planQuery.isLoading && (
                    <div className="plan-diff plan-diff-empty">
                      Loading plan…
                    </div>
                  )}
                  {planQuery.isError && (
                    <div className="plan-diff plan-diff-empty">
                      Full plan unavailable — showing the resource summary only.
                    </div>
                  )}
                  {planQuery.data &&
                    (planResource ? (
                      <PlanResourceDiff
                        resource={planResource}
                        action={resource.action}
                      />
                    ) : (
                      <div className="plan-diff plan-diff-empty">
                        Resource not found in the recorded plan.
                      </div>
                    ))}
                </div>
              )}
            </li>
          );
        })}
      </ul>
      <div className="drift-resources-foot">
        {drift.resources_truncated && (
          <span>
            Resource list truncated — download plan.json for the full plan
          </span>
        )}
        {jobID && (
          <a
            className="link"
            href={planDownloadURL(jobID)}
            download={`${jobID}-plan.json`}
          >
            <Download style={{ width: 13, height: 13 }} /> plan.json
          </a>
        )}
      </div>
    </div>
  );
};

/* Sandbox */

const formatDuration = (start?: string, end?: string) => {
  if (!start || !end) return "—";
  const seconds = Math.max(
    0,
    Math.round((Date.parse(end) - Date.parse(start)) / 1000),
  );
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
};

const displayValue = (value: unknown) =>
  typeof value === "string" ? value : JSON.stringify(value, null, 2);

type TSandboxRow = {
  id: string;
  name: string;
  status: string;
  operation?: string;
  started_at?: string;
  finished_at?: string;
  error?: string;
  executions?: number;
  outputs?: unknown;
};

const SandboxPage = () => {
  const report = useQuery({
    queryKey: ["report"],
    queryFn: getReport,
    retry: false,
  });
  const status = useQuery({
    queryKey: ["status"],
    queryFn: getStatus,
    retry: false,
  });
  const runner = useRunnerState();
  const [expanded, setExpanded] = useState<string | null>(null);

  const reportRows: TSandboxRow[] = (report.data?.steps ?? [])
    .filter((step) => step.job_group === "sandbox")
    .map((step) => ({
      id: step.id,
      name: step.name,
      status: step.status,
      operation: step.job_operation,
      started_at: step.started_at,
      finished_at: step.finished_at,
      error: step.error,
      executions: step.executions,
      outputs: step.outputs,
    }));
  const fallbackRows: TSandboxRow[] = (status.data?.steps ?? [])
    .filter((step) =>
      `${step.id} ${step.name}`.toLowerCase().includes("sandbox"),
    )
    .map((step) => ({ ...step }));
  const rows = reportRows.length > 0 ? reportRows : fallbackRows;

  const latest = rows[rows.length - 1];
  const typeMatch = latest?.name.match(/\(([^)]+)\)/);
  const sandboxType = typeMatch?.[1] ?? (rows.length > 0 ? "sandbox" : "—");
  const loading = report.isLoading || status.isLoading;

  return (
    <main>
      <PageHeader
        title="Sandbox"
        subtitle="Provisioning status for the sandbox this bundle installs into."
      />

      <div className="tiles">
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Provision</span>
            {latest ? <Badge>{latest.status}</Badge> : <Clock />}
          </div>
          <div className="tile-value">{latest?.status ?? "unknown"}</div>
          <div className="tile-note">
            {latest?.finished_at
              ? `finished ${formatTime(latest.finished_at)}`
              : latest?.started_at
                ? `started ${formatTime(latest.started_at)}`
                : "no sandbox step recorded"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Type</span>
            <Cpu />
          </div>
          <div className="tile-value">{sandboxType}</div>
          <div className="tile-note">
            {latest ? latest.name : "from the bundle envelope"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Duration</span>
            <Clock />
          </div>
          <div className="tile-value">
            {formatDuration(latest?.started_at, latest?.finished_at)}
          </div>
          <div className="tile-note">
            {latest?.executions
              ? `${latest.executions} execution${latest.executions === 1 ? "" : "s"}`
              : "provision time"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Runner</span>
            <CircleDot
              className={
                runner.state === "active"
                  ? "good"
                  : runner.state === "offline"
                    ? "warn"
                    : ""
              }
            />
          </div>
          <div className="tile-value">{runner.state}</div>
          <div className="tile-note">
            last observed {formatTime(runner.observedAt)}
          </div>
        </div>
      </div>

      <section className="card">
        <div className="card-header">
          <h3>
            <Cpu /> Sandbox steps
          </h3>
          {report.data && <Badge mono>run {report.data.run_id}</Badge>}
        </div>
        <State
          loading={loading}
          empty={!loading && rows.length === 0}
          emptyText="No sandbox steps found — the airgapped runner records them during bootstrap"
        />
        {rows.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Step</th>
                  <th>Status</th>
                  <th>Operation</th>
                  <th>Started</th>
                  <th>Finished</th>
                  <th>Logs</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => {
                  const expandable = row.outputs != null;
                  const isExpanded = expandable && expanded === row.id;
                  return (
                    <Fragment key={row.id}>
                      <tr
                        className={
                          expandable
                            ? `selectable ${isExpanded ? "selected" : ""}`
                            : undefined
                        }
                        onClick={
                          expandable
                            ? () => setExpanded(isExpanded ? null : row.id)
                            : undefined
                        }
                      >
                        <td>
                          <span className="strong">
                            {expandable ? (isExpanded ? "▾ " : "▸ ") : ""}
                            {row.name}
                          </span>
                          <code className="subtle">{row.id}</code>
                        </td>
                        <td>
                          <Badge>{row.status}</Badge>
                          {row.error && (
                            <div
                              style={{
                                fontSize: 12,
                                color: "var(--error-text)",
                                marginTop: 4,
                              }}
                            >
                              {row.error}
                            </div>
                          )}
                        </td>
                        <td
                          style={{
                            fontFamily: "var(--font-mono)",
                            fontSize: 12,
                          }}
                        >
                          {row.operation ?? "—"}
                        </td>
                        <td style={{ fontSize: 12, color: "var(--muted)" }}>
                          {formatTime(row.started_at)}
                        </td>
                        <td style={{ fontSize: 12, color: "var(--muted)" }}>
                          {formatTime(row.finished_at)}
                        </td>
                        <td onClick={(event) => event.stopPropagation()}>
                          <Link
                            className="link"
                            to={`/logs?job=${encodeURIComponent(row.id)}`}
                          >
                            logs →
                          </Link>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr>
                          <td colSpan={6}>
                            <pre
                              style={{
                                margin: 0,
                                padding: "8px 0",
                                fontSize: 12,
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                            >
                              {displayValue(row.outputs)}
                            </pre>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
};

/* Stack */

const StackPage = () => {
  const outputs = useQuery({
    queryKey: ["stack-outputs"],
    queryFn: getStackOutputs,
    retry: false,
  });
  const status = useQuery({
    queryKey: ["status"],
    queryFn: getStatus,
    retry: false,
  });
  const runner = useRunnerState();

  const entries = Object.entries(outputs.data ?? {}).sort(([a], [b]) =>
    a.localeCompare(b),
  );
  const published = outputs.isSuccess;
  const missing =
    outputs.isError &&
    outputs.error.message.toLowerCase().includes("not found");

  return (
    <main>
      <PageHeader
        title="Stack"
        subtitle="Install stack readiness and the outputs the airgapped runner late-binds into every plan."
      />

      <div className="tiles">
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Outputs</span>
            {published ? (
              <CheckCircle2 className="good" />
            ) : (
              <Clock className={missing ? "warn" : ""} />
            )}
          </div>
          <div className="tile-value">
            {published
              ? "published"
              : missing
                ? "waiting"
                : outputs.isLoading
                  ? "…"
                  : "error"}
          </div>
          <div className="tile-note">
            {published
              ? "phone-home outputs available"
              : missing
                ? "waiting for the stack phone-home"
                : outputs.isLoading
                  ? "loading"
                  : outputs.error?.message}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Keys</span>
            <Layers />
          </div>
          <div className="tile-value">{published ? entries.length : "—"}</div>
          <div className="tile-note">stack output values</div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Install</span>
            <Package />
          </div>
          <div
            className="tile-value"
            style={{
              fontSize: 16,
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {status.data?.install_id ?? "—"}
          </div>
          <div className="tile-note">
            {status.data?.heartbeat_at
              ? `heartbeat ${formatTime(status.data.heartbeat_at)}`
              : "no install status published"}
          </div>
        </div>
        <div className="tile">
          <div className="tile-top">
            <span className="tile-label">Runner</span>
            <CircleDot
              className={
                runner.state === "active"
                  ? "good"
                  : runner.state === "offline"
                    ? "warn"
                    : ""
              }
            />
          </div>
          <div className="tile-value">{runner.state}</div>
          <div className="tile-note">
            last observed {formatTime(runner.observedAt)}
          </div>
        </div>
      </div>

      <section className="card">
        <div className="card-header">
          <h3>
            <Layers /> Stack outputs
          </h3>
        </div>
        <State
          error={outputs.isError && !missing ? outputs.error : undefined}
          loading={outputs.isLoading}
          empty={!outputs.isLoading && !outputs.isError && entries.length === 0}
          emptyText="The stack published an empty output document"
        />
        {missing && (
          <div className="state">
            Stack outputs are not published yet. The install stack
            (CloudFormation) writes them via its phone-home Lambda once
            provisioning finishes; the airgapped runner polls for the same
            document before it runs any plan.
          </div>
        )}
        {entries.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Output</th>
                  <th>Value</th>
                </tr>
              </thead>
              <tbody>
                {entries.map(([key, value]) => (
                  <tr key={key}>
                    <td>
                      <span className="strong">{key}</span>
                    </td>
                    <td>
                      <pre
                        style={{
                          margin: 0,
                          fontFamily: "var(--font-mono)",
                          fontSize: 12,
                          whiteSpace: "pre-wrap",
                          wordBreak: "break-all",
                        }}
                      >
                        {displayValue(value)}
                      </pre>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
};

/* Bundles */

const formatBytes = (value?: number) => {
  if (!value) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size >= 10 || unit === 0 ? Math.round(size) : size.toFixed(1)} ${units[unit]}`;
};

const CONTENT_KINDS = [
  { kind: "component", label: "Components", Icon: Box },
  { kind: "sandbox", label: "Sandbox", Icon: Server },
  { kind: "image", label: "Images", Icon: Container },
  { kind: "action", label: "Actions", Icon: Zap },
  { kind: "runbook", label: "Runbooks", Icon: BookOpen },
  { kind: "stack-asset", label: "Stack assets", Icon: FileCode },
  { kind: "runner-binary", label: "Runner", Icon: Cpu },
  { kind: "runner-image", label: "Runner image", Icon: Container },
];

const ContentKindIcon = ({ kind }: { kind: string }) => {
  const Icon = CONTENT_KINDS.find((k) => k.kind === kind)?.Icon ?? Layers;
  return <Icon />;
};

const contentSummary = (contents: TBundleContent[]) => {
  const counts = new Map<string, number>();
  for (const item of contents)
    counts.set(item.kind, (counts.get(item.kind) ?? 0) + 1);
  return CONTENT_KINDS.filter((k) => counts.has(k.kind))
    .map((k) => `${counts.get(k.kind)} ${k.label.toLowerCase()}`)
    .join(" · ");
};

const VerificationRow = ({
  ok,
  label,
  children,
}: {
  ok: boolean;
  label: string;
  children?: ReactNode;
}) => (
  <div className="verify-row">
    {ok ? <CheckCircle2 className="ok" /> : <XCircle className="fail" />}
    <div>
      <div>{label}</div>
      {children && <div className="verify-sub">{children}</div>}
    </div>
  </div>
);

const HistoryEntry = ({
  info,
  active,
  comparison,
  upgradeRun,
  stackCandidate,
  expanded,
  onToggle,
  hideHeader = false,
  hideDetails = false,
}: {
  info: TBundleInfo;
  active: boolean;
  comparison?: TBundleHistoryComparison;
  upgradeRun?: TRun;
  stackCandidate?: TStackCandidate;
  expanded: boolean;
  onToggle: () => void;
  hideHeader?: boolean;
  hideDetails?: boolean;
}) => {
  const [expandedPlan, setExpandedPlan] = useState<string | null>(null);
  const changePriority: Record<string, number> = {
    component: 0,
    sandbox: 1,
    "stack-asset": 2,
    action: 3,
    "runner-binary": 4,
    "runner-image": 5,
  };
  const changes = (comparison?.changes ?? [])
    .filter((change) => change.change !== "unchanged")
    .sort(
      (left, right) =>
        (changePriority[left.kind] ?? 99) -
          (changePriority[right.kind] ?? 99) ||
        left.name.localeCompare(right.name),
    );
  const counts = changes.reduce(
    (summary, change) => ({
      ...summary,
      [change.change]: summary[change.change] + 1,
    }),
    { changed: 0, added: 0, removed: 0, unchanged: 0 },
  );
  const groups = [
    {
      label: "Components",
      description: "Application workloads",
      changes: changes.filter((change) => change.kind === "component"),
    },
    {
      label: "Infrastructure",
      description: "Sandbox and install stack",
      changes: changes.filter(
        (change) => change.kind === "sandbox" || change.kind === "stack-asset",
      ),
    },
    {
      label: "Operations",
      description: "Actions and runbooks",
      changes: changes.filter((change) => change.kind === "action" || change.kind === "runbook"),
    },
    {
      label: "Packaging",
      description: "Runner and supporting artifacts",
      changes: changes.filter(
        (change) =>
          change.kind !== "component" &&
          change.kind !== "sandbox" &&
          change.kind !== "stack-asset" &&
          change.kind !== "action",
      ),
    },
  ].filter((group) => group.changes.length > 0);
  const impactSummary = groups
    .map((group) => `${group.changes.length} ${group.label.toLowerCase()}`)
    .join(" · ");
  const planStep = (change: TBundleChange): TStatusStep | null => {
    if (!active || !change.plan_step_id || !upgradeRun) return null;
    const step = upgradeRun.steps?.find(
      (candidate) => candidate.id === change.plan_step_id,
    );
    if (!step || step.status !== "finished") return null;
    return {
      id: step.id,
      name: step.name,
      status: step.status,
      started_at: step.started_at,
      finished_at: step.finished_at,
    };
  };
  return (
    <div className={`history-activation ${hideHeader ? "history-detail" : ""}`}>
      {!hideHeader && <div className="activity-item">
        <FileArchive
          style={{ width: 14, height: 14, flexShrink: 0, color: "var(--muted)" }}
        />
        <span className="activity-body">
          <span style={{ fontFamily: "var(--font-mono)", fontSize: 12 }}>
            {info.bundle_digest.replace("sha256:", "").slice(0, 16)}…
          </span>
          <code>
            activated {formatTime(info.activated_at)}
            {comparison?.available && impactSummary ? ` · ${impactSummary}` : ""}
          </code>
        </span>
        {comparison && (
          <button
            type="button"
            className={`chip ${expanded ? "on" : ""}`}
            disabled={!comparison.available}
            onClick={onToggle}
          >
            <ChevronRight className={expanded ? "expanded-chevron" : ""} />
            {comparison.available
              ? "View changes"
              : "diff unavailable"}
          </button>
        )}
        <Badge theme={active ? "success" : ""} mono>
          {active ? "active" : "superseded"}
        </Badge>
      </div>}
      {!hideDetails && expanded && comparison?.available && (
        <div className="history-diff">
          <div className="history-diff-heading">
            <span>Bundle changes from</span>
            <code>{shortDigest(comparison.previous_digest)}</code>
            <span>to</span>
            <code>{shortDigest(comparison.bundle_digest)}</code>
          </div>
          <div className="history-change-summary">
            {counts.changed > 0 && <Badge theme="info">{counts.changed} changed</Badge>}
            {counts.added > 0 && <Badge theme="success">{counts.added} added</Badge>}
            {counts.removed > 0 && <Badge theme="error">{counts.removed} removed</Badge>}
          </div>
          {changes.length === 0 ? (
            <div className="notice">No bundle inventory changes.</div>
          ) : (
            <div className="history-change-groups">
              {groups.map((group) => (
                <section className="history-change-group" key={group.label}>
                  <div className="history-group-heading">
                    <div>
                      <strong>{group.label}</strong>
                      <span>{group.description}</span>
                    </div>
                    <Badge mono>{group.changes.length}</Badge>
                  </div>
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr><th>Name</th><th>Change</th><th>What changed</th><th>Review</th></tr>
                      </thead>
                      <tbody>
                        {group.changes.map((change, index) => {
                          const key = `${change.kind}-${change.name}-${index}`;
                          const step = planStep(change);
                          const stackDiff =
                            change.kind === "stack-asset" &&
                            stackCandidate?.bundle_digest === info.bundle_digest
                              ? stackCandidate
                              : null;
                          const actionDiff =
                            change.kind === "action" &&
                            (change.previous_action_definition || change.candidate_action_definition);
                          const componentDiff =
                            change.kind === "component" &&
                            (change.previous_component_definition || change.candidate_component_definition);
                          const runbookDiff =
                            change.kind === "runbook" &&
                            (change.previous_runbook_definition || change.candidate_runbook_definition);
                          const planExpanded = expandedPlan === key;
                          return (
                            <Fragment key={key}>
                              <tr>
                                <td><span className="name-cell"><ContentKindIcon kind={change.kind} /><span><span className="strong">{change.name}</span><code className="subtle">{change.kind}{change.detail ? ` · ${change.detail}` : ""}</code></span></span></td>
                                <td><Badge theme={change.change === "removed" ? "error" : change.change === "added" ? "success" : "info"} mono>{change.change}</Badge></td>
                                <td><BundleChangeDescription change={change} /></td>
                                <td>
                                  {step || stackDiff || componentDiff || actionDiff || runbookDiff ? (
                                    <button type="button" className="link history-plan-button" onClick={() => setExpandedPlan(planExpanded ? null : key)}>
                                      <ChevronRight className={planExpanded ? "expanded-chevron" : ""} />
                                      {planExpanded
                                        ? "Hide diff"
                                        : actionDiff
                                          ? "View bundled definition"
                                          : runbookDiff
                                            ? "View bundled runbook"
                                            : componentDiff && step
                                              ? "View definition & deployed diff"
                                              : componentDiff
                                                ? "View bundled definition"
                                          : stackDiff
                                          ? "View CloudFormation diff"
                                          : "View deployed diff"}
                                    </button>
                                  ) : "—"}
                                </td>
                              </tr>
                              {planExpanded && (step || stackDiff || componentDiff || actionDiff || runbookDiff) && (
                                <tr className="candidate-plan-row"><td colSpan={4}>
                                  {actionDiff
                                    ? <ActionDefinitionDiff before={change.previous_action_definition} after={change.candidate_action_definition} />
                                    : runbookDiff
                                      ? <RunbookDefinitionDiff before={change.previous_runbook_definition} after={change.candidate_runbook_definition} />
                                    : componentDiff
                                      ? <><ComponentDefinitionDiff before={change.previous_component_definition} after={change.candidate_component_definition} />{step && <StepPlanPreview step={step} />}</>
                                    : step
                                      ? <StepPlanPreview step={step} />
                                      : stackDiff
                                        ? <StackChangesTable candidate={stackDiff} />
                                        : null}
                                </td></tr>
                              )}
                            </Fragment>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </section>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const StackChangesTable = ({ candidate, heading = true }: { candidate: TStackCandidate; heading?: boolean }) => {
  const changes = candidate.changes ?? [];
  const dependencyReevaluations = changes.filter(
    (change) => change.action === "Modify" && change.property_changes_captured && (change.property_changes ?? []).length === 0,
  );
  const materialChanges = changes.filter((change) => !dependencyReevaluations.includes(change));
  const rows = (items: TStackChange[]) => items.map((change, index) => {
    const key = `${change.logical_resource_id}-${index}`;
    const propertyChanges = change.property_changes ?? [];
    return (
      <Fragment key={key}>
        <tr>
          <td><span className="strong">{change.logical_resource_id}</span></td>
          <td><code>{change.resource_type}</code></td>
          <td><Badge theme={change.action === "Remove" ? "error" : change.action === "Add" ? "success" : "info"} mono>{change.action}</Badge></td>
          <td>{change.replacement || "—"}</td>
          <td>{change.details?.map((detail) => detail.name || detail.attribute).filter(Boolean).join(", ") || change.scope?.join(", ") || "—"}</td>
          <td>{propertyChanges.length > 0 ? `${propertyChanges.length} value ${propertyChanges.length === 1 ? "change" : "changes"}` : "evaluation only"}</td>
        </tr>
        <tr className="stack-property-row">
          <td colSpan={6}><StackResourceChanges change={change} /></td>
        </tr>
      </Fragment>
    );
  });
  const table = (items: TStackChange[]) => (
    <div className="table-wrap">
      <table>
        <thead><tr><th>Logical resource</th><th>Type</th><th>Action</th><th>Replacement</th><th>Changed properties</th><th>Values</th></tr></thead>
        <tbody>{rows(items)}</tbody>
      </table>
    </div>
  );
  return (
    <div>
      {heading && <div className="history-stack-heading">
        <span>CloudFormation change set</span>
        <code>{candidate.change_set_name}</code>
        <Badge theme={candidate.stack_applied_at ? "success" : "info"} mono>
          {candidate.stack_applied_at ? "applied" : "ready"}
        </Badge>
      </div>}
      {materialChanges.length > 0
        ? table(materialChanges)
        : <div className="notice candidate-notice">No template property values change in this plan.</div>}
      {dependencyReevaluations.length > 0 && (
        <details className="stack-reevaluations">
          <summary>
            <span><strong>Dependency reevaluations</strong><small>CloudFormation revisited these resources, but their template property values do not change.</small></span>
            <span><Badge>{dependencyReevaluations.length}</Badge><ChevronRight /></span>
          </summary>
          {table(dependencyReevaluations)}
        </details>
      )}
    </div>
  );
};

const StackResourceChanges = ({ change }: { change: TStackChange }) => (
  <div className="stack-resource-detail">
    {(change.property_changes ?? []).length > 0 ? (
      <StackPropertyChanges changes={change.property_changes ?? []} truncated={change.property_changes_truncated} />
    ) : (
      <div className="notice stack-values-unavailable">
        {change.property_changes_captured
          ? "No template property value changed. CloudFormation reevaluated this resource because of the dependency shown below."
          : "Exact before/after values were not captured for this historical change set. CloudFormation’s recorded evaluation details are shown below."}
      </div>
    )}
    {(change.details ?? []).length > 0 && (
      <div className="stack-change-reasons">
        <div className="stack-change-reasons-heading">CloudFormation evaluation</div>
        {(change.details ?? []).map((detail, index) => (
          <div className="stack-change-reason" key={`${detail.attribute}-${detail.name}-${index}`}>
            <code>{detail.name || detail.attribute || "resource"}</code>
            <span>{detail.change_source || "change"}</span>
            <span>{detail.evaluation || "—"}</span>
            <span>{detail.requires_recreation || "—"}</span>
            <code>{detail.causing_entity || "—"}</code>
          </div>
        ))}
      </div>
    )}
  </div>
);

const StackPropertyChanges = ({ changes, truncated }: { changes: TStackPropertyChange[]; truncated?: boolean }) => (
  <div className="stack-property-diff">
    <div className="stack-property-header">Changed properties</div>
    {changes.map((change) => (
      <section className="stack-property-change" key={change.path}>
        <code className="stack-property-name">{change.path.replace(/^Properties\./, "")}</code>
        <ValueDiff before={change.before} after={change.after} />
      </section>
    ))}
    {truncated && <div className="notice">Additional property changes were omitted from the persisted comparison.</div>}
  </div>
);

const ActionDefinitionDiff = ({ before, after }: { before?: TBundleActionDefinition; after?: TBundleActionDefinition }) => {
  const beforeSteps = new Map((before?.steps ?? []).map((step) => [step.name, step]));
  const afterSteps = new Map((after?.steps ?? []).map((step) => [step.name, step]));
  const names = Array.from(new Set([...Array.from(beforeSteps.keys()), ...Array.from(afterSteps.keys())])).filter(
    (name) => JSON.stringify(beforeSteps.get(name)) !== JSON.stringify(afterSteps.get(name)),
  );
  const { steps: _beforeSteps, ...beforeSettings } = before ?? {};
  const { steps: _afterSteps, ...afterSettings } = after ?? {};
  const settingChanges = definitionPropertyChanges(beforeSettings, afterSettings);
  return (
    <div className="action-definition-diff">
      <div className="history-stack-heading">
        <span>Bundled action definition</span>
        <Badge mono>{settingChanges.length} settings · {names.length} steps changed</Badge>
      </div>
      {settingChanges.length > 0 && <StackPropertyChanges changes={settingChanges} />}
      {names.length > 0 ? (
        names.map((name) => (
          <ActionStepDiff key={name} name={name} before={beforeSteps.get(name)} after={afterSteps.get(name)} />
        ))
      ) : settingChanges.length === 0 ? (
        <div className="notice">The canonical action definition is unchanged.</div>
      ) : null}
    </div>
  );
};

const ActionStepDiff = ({ name, before, after }: { name: string; before?: TBundleActionStep; after?: TBundleActionStep }) => {
  const values = (step?: TBundleActionStep): Record<string, unknown> => ({
    command: step?.command,
    "inline contents": step?.inline_contents_digest,
    repository: step?.source?.repository,
    ref: step?.source?.requested_ref,
    commit: step?.source?.commit,
    directory: step?.source?.directory,
    version: step?.source?.version,
    "source digest": step?.source?.digest,
    "artifact digest": step?.artifact_digest,
    index: step?.index,
    environment: step?.environment,
  });
  const beforeValues = values(before);
  const afterValues = values(after);
  const fields = Object.keys(beforeValues).filter(
    (field) => beforeValues[field] !== afterValues[field],
  );
  const change = !before ? "added" : !after ? "removed" : "changed";
  return (
    <section className="action-step-diff">
      <div className="action-step-heading">
        <code>{name}</code>
        <Badge theme={change === "added" ? "success" : change === "removed" ? "error" : "info"} mono>{change}</Badge>
      </div>
      {fields.length > 0 ? (
        <div className="stack-property-diff">
          <div className="stack-property-header">Changed fields</div>
          {fields.map((field) => (
            <section className="stack-property-change" key={field}>
              <code className="stack-property-name">{field}</code>
              <ValueDiff before={beforeValues[field]} after={afterValues[field]} />
            </section>
          ))}
        </div>
      ) : (
        <div className="notice">The persisted step definition is unchanged.</div>
      )}
    </section>
  );
};

const RunbookDefinitionDiff = ({ before, after }: { before?: TBundleRunbookDefinition; after?: TBundleRunbookDefinition }) => {
  const changes = definitionPropertyChanges(before, after);
  return (
    <div className="action-definition-diff">
      <div className="history-stack-heading">
        <span>Bundled runbook definition</span>
        <Badge mono>{changes.length} changed {changes.length === 1 ? "field" : "fields"}</Badge>
      </div>
      {changes.length > 0
        ? <StackPropertyChanges changes={changes} />
        : <div className="notice">The canonical runbook definition is unchanged.</div>}
    </div>
  );
};

const ComponentDefinitionDiff = ({ before, after }: { before?: Record<string, unknown>; after?: Record<string, unknown> }) => {
  const changes = definitionPropertyChanges(before, after);
  const beforeToml = before ? componentDefinitionToml(before) : null;
  const afterToml = after ? componentDefinitionToml(after) : null;
  const comparisonAvailable = beforeToml !== null && afterToml !== null;
  const payloads = comparisonAvailable
    ? lineDiff(beforeToml, afterToml)
    : (afterToml ?? beforeToml ?? "").split("\n").map((line) => ({ line, type: "context" }));
  return (
    <div className="action-definition-diff">
      <div className="history-stack-heading">
        <span>Bundled component definition</span>
        <Badge mono>{comparisonAvailable ? `${changes.length} changed ${changes.length === 1 ? "field" : "fields"}` : "candidate snapshot"}</Badge>
      </div>
      {!comparisonAvailable && payloads.length > 0 && (
        <div className="notice">The active bundle predates canonical component definitions. Showing the uploaded TOML snapshot; a line diff is unavailable for this upgrade.</div>
      )}
      {payloads.length > 0
        ? <ManifestLineDiff payloads={payloads} changedOnly={comparisonAvailable} />
        : <div className="notice">The canonical component definition is unchanged.</div>}
    </div>
  );
};

const tomlKey = (key: string) => /^[A-Za-z0-9_-]+$/.test(key) ? key : JSON.stringify(key);

const tomlScalar = (value: unknown): string => {
  if (typeof value === "string") {
    if (value.includes("\n") && !value.includes("'''")) return `'''\n${value.replace(/\n$/, "")}\n'''`;
    return JSON.stringify(value);
  }
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  if (Array.isArray(value)) return `[${value.map(tomlScalar).join(", ")}]`;
  if (value && typeof value === "object") {
    return `{ ${Object.entries(value).filter(([, child]) => child != null).map(([key, child]) => `${tomlKey(key)} = ${tomlScalar(child)}`).join(", ")} }`;
  }
  return "";
};

const componentDefinitionToml = (definition: Record<string, unknown>): string => {
  const lines: string[] = [];
  const writeTable = (value: Record<string, unknown>, path: string[]) => {
    const entries = Object.entries(value).filter(([, child]) => child != null);
    const scalars = entries.filter(([, child]) => typeof child !== "object" || Array.isArray(child));
    const tables = entries.filter(([, child]) => child && typeof child === "object" && !Array.isArray(child) && Object.keys(child).length > 0);
    const emptyTables = entries.filter(([, child]) => child && typeof child === "object" && !Array.isArray(child) && Object.keys(child).length === 0);
    if (path.length > 0) {
      if (lines.length > 0) lines.push("");
      lines.push(`[${path.map(tomlKey).join(".")}]`);
    }
    [...scalars, ...emptyTables].sort(([left], [right]) => left.localeCompare(right)).forEach(([key, child]) => {
      lines.push(`${tomlKey(key)} = ${child && typeof child === "object" && !Array.isArray(child) ? "{}" : tomlScalar(child)}`);
    });
    tables.sort(([left], [right]) => left.localeCompare(right)).forEach(([key, child]) => {
      writeTable(child as Record<string, unknown>, [...path, key]);
    });
  };
  writeTable(definition, []);
  return lines.join("\n");
};

const lineDiff = (before: string, after: string): Array<{ line: string; type: string }> => {
  const oldLines = before.split("\n");
  const newLines = after.split("\n");
  const lengths = Array.from({ length: oldLines.length + 1 }, () => new Array<number>(newLines.length + 1).fill(0));
  for (let oldIndex = oldLines.length - 1; oldIndex >= 0; oldIndex--) {
    for (let newIndex = newLines.length - 1; newIndex >= 0; newIndex--) {
      lengths[oldIndex][newIndex] = oldLines[oldIndex] === newLines[newIndex]
        ? lengths[oldIndex + 1][newIndex + 1] + 1
        : Math.max(lengths[oldIndex + 1][newIndex], lengths[oldIndex][newIndex + 1]);
    }
  }
  const payloads: Array<{ line: string; type: string }> = [];
  let oldIndex = 0;
  let newIndex = 0;
  while (oldIndex < oldLines.length && newIndex < newLines.length) {
    if (oldLines[oldIndex] === newLines[newIndex]) {
      payloads.push({ line: oldLines[oldIndex], type: "context" });
      oldIndex++;
      newIndex++;
    } else if (lengths[oldIndex + 1][newIndex] >= lengths[oldIndex][newIndex + 1]) {
      payloads.push({ line: oldLines[oldIndex++], type: "removed" });
    } else {
      payloads.push({ line: newLines[newIndex++], type: "added" });
    }
  }
  while (oldIndex < oldLines.length) payloads.push({ line: oldLines[oldIndex++], type: "removed" });
  while (newIndex < newLines.length) payloads.push({ line: newLines[newIndex++], type: "added" });
  return payloads;
};

const definitionPropertyChanges = (before: unknown, after: unknown, path = "definition"): TStackPropertyChange[] => {
  if (JSON.stringify(before) === JSON.stringify(after)) return [];
  if (before && after && typeof before === "object" && typeof after === "object" && !Array.isArray(before) && !Array.isArray(after)) {
    const oldValues = before as Record<string, unknown>;
    const newValues = after as Record<string, unknown>;
    return Array.from(new Set([...Object.keys(oldValues), ...Object.keys(newValues)]))
      .sort()
      .flatMap((key) => definitionPropertyChanges(oldValues[key], newValues[key], `${path}.${key}`));
  }
  return [{ path, before, after }];
};

const BundleChangeDescription = ({ change }: { change: TBundleChange }) => {
  if (change.change === "added") return <span>Added to the bundle</span>;
  if (change.change === "removed") return <span>Removed from the bundle</span>;
  const actionStepsUnchanged =
    change.kind === "action" &&
    change.previous_action_definition &&
    change.candidate_action_definition &&
    JSON.stringify(change.previous_action_definition) === JSON.stringify(change.candidate_action_definition);
  const contentChanged = change.previous_digest !== change.candidate_digest;
  const configChanged =
    change.previous_config_digest !== change.candidate_config_digest;
  const configDuplicatesContent =
    change.previous_digest === change.previous_config_digest &&
    change.candidate_digest === change.candidate_config_digest;
  const subject =
    change.kind === "component"
      ? "Component"
      : change.kind === "sandbox"
        ? "Sandbox Terraform"
        : change.kind === "action"
          ? "Action workflow"
          : change.kind === "runbook"
            ? "Runbook"
          : change.kind === "stack-asset"
            ? "CloudFormation template"
            : change.kind === "runner-binary"
              ? "Runner binary"
              : change.kind === "runner-image"
                ? "Runner image"
                : "Bundle content";
  const detail =
    actionStepsUnchanged
      ? "bundle metadata changed; executable steps are unchanged"
      : contentChanged && configChanged
      ? "content and configuration changed"
      : configChanged
        ? "configuration changed"
        : contentChanged
          ? "content changed"
          : "metadata changed";
  return (
    <span className="bundle-change-description">
      <span>{subject} {detail}</span>
      <span className="digest-evidence">
        {contentChanged && <span>{actionStepsUnchanged ? "bundle metadata" : configDuplicatesContent ? "definition" : "content"} <DigestChange before={change.previous_digest} after={change.candidate_digest} /></span>}
        {configChanged && !configDuplicatesContent && <span>config <DigestChange before={change.previous_config_digest} after={change.candidate_config_digest} /></span>}
      </span>
    </span>
  );
};

const shortDigest = (digest?: string) =>
  digest?.replace("sha256:", "").slice(0, 12) ?? "—";

const DigestChange = ({ before, after }: { before?: string; after?: string }) => (
  <span className="digest-change" title={[before, after].filter(Boolean).join(" → ")}>
    {before && before !== after && <code>{shortDigest(before)}</code>}
    {before && before !== after && <span>→</span>}
    <code>{shortDigest(after)}</code>
  </span>
);

const Bundles = () => {
  const queryClient = useQueryClient();
  const bundle = useQuery({
    queryKey: ["bundle"],
    queryFn: getBundle,
    refetchInterval: (query) => {
      const data = query.state.data;
      if (data?.stack_candidate?.status === "CREATE_PENDING") return 2000;
      if (data?.candidate && !data.stack_candidate?.runner_activated_at) return 5000;
      return false;
    },
  });
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const status = useQuery({
    queryKey: ["status"],
    queryFn: getStatus,
    retry: false,
    refetchInterval: 5000,
  });
  const [hiddenKinds, setHiddenKinds] = useState<Set<string>>(new Set());
  const [showUnchanged, setShowUnchanged] = useState(false);
  const [openChangeGroups, setOpenChangeGroups] = useState<string[]>(["Components"]);
  const [expandedChange, setExpandedChange] = useState<string | null>(null);
  const [expandedHistory, setExpandedHistory] = useState<string | null>(null);
  const [bundleView, setBundleView] = useState<"review" | "active" | "history">("review");
  const [reviewView, setReviewView] = useState<"changes" | "plan">("changes");
  const [showBundleUpload, setShowBundleUpload] = useState(false);
  const [showClearCandidate, setShowClearCandidate] = useState(false);
  const [applyCommandCopied, setApplyCommandCopied] = useState(false);
  const [draggingBundle, setDraggingBundle] = useState(false);
  const [bundleUploadProgress, setBundleUploadProgress] = useState({ loaded: 0, total: 0 });
  const approve = useMutation({
    mutationFn: approveBundleCandidate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bundle"] });
      queryClient.invalidateQueries({ queryKey: ["status"] });
      queryClient.invalidateQueries({ queryKey: ["step-result"] });
      queryClient.invalidateQueries({ queryKey: ["step-plan"] });
    },
  });
  const clearCandidate = useMutation({
    mutationFn: clearBundleCandidate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bundle"] });
      setExpandedChange(null);
      setShowClearCandidate(false);
      setReviewView("changes");
    },
  });
  const planStack = useMutation({
    mutationFn: planBundleCandidateStack,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bundle"] });
      queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
  });
  const uploadBundle = useMutation({
    mutationFn: (file: File) => uploadBundleCandidate(file, (loaded, total) => setBundleUploadProgress({ loaded, total })),
    onMutate: () => {
      queryClient.removeQueries({ queryKey: ["bundle-upload-status"] });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bundle"] });
      setExpandedChange(null);
      setShowBundleUpload(false);
      setBundleView("review");
      setReviewView("changes");
    },
  });
  const bundleUploadStatus = useQuery({
    queryKey: ["bundle-upload-status"],
    queryFn: getBundleUploadStatus,
    enabled: uploadBundle.isPending && bundleUploadProgress.total > 0 && bundleUploadProgress.loaded >= bundleUploadProgress.total,
    refetchInterval: 750,
  });
  const selectBundle = (files: FileList | null) => {
    const file = files?.[0];
    if (file) {
      setBundleUploadProgress({ loaded: 0, total: file.size });
      uploadBundle.mutate(file);
    }
  };

  const active = bundle.data?.active;
  const candidate =
    bundle.data?.candidate?.bundle.bundle_digest === active?.bundle_digest
      ? null
      : bundle.data?.candidate;
  const candidateRecordKey = bundle.data?.candidate_record_key;
  const stagedStack = bundle.data?.stack_candidate;
  const stackCandidate =
    candidateRecordKey && stagedStack?.candidate_record_key === candidateRecordKey
      ? stagedStack
      : null;
  const stackPlanRunning = stackCandidate?.status === "CREATE_PENDING";
  const stackPlanFailed = stackCandidate?.status === "FAILED";
  const stackPending = !!stackCandidate && !stackPlanRunning && !stackPlanFailed && !stackCandidate.runner_activated_at;
  const history = bundle.data?.history ?? [];
  const candidatePlanRun = runs.data?.find((run) => run.ref_kind === "bundle-plan" && run.ref_id === candidateRecordKey);
  const candidatePlanSteps = candidatePlanRun?.steps ?? [];
  const historyComparisons = bundle.data?.comparisons ?? [];
  useEffect(() => {
    if (
      active?.bundle_digest &&
      historyComparisons.some(
        (comparison) =>
          comparison.bundle_digest === active.bundle_digest &&
          comparison.available,
      )
    ) {
      setExpandedHistory(active.bundle_digest);
    }
  }, [active?.bundle_digest]);
  const contents = active?.contents ?? [];
  const presentKinds = CONTENT_KINDS.filter((k) =>
    contents.some((item) => item.kind === k.kind),
  );
  const visible = contents.filter((item) => !hiddenKinds.has(item.kind));
  const digest = (
    active?.bundle_digest ?? catalog.data?.bundle_digest
  )?.replace("sha256:", "");

  const toggleKind = (kind: string) =>
    setHiddenKinds((prev) => {
      const next = new Set(prev);
      if (next.has(kind)) next.delete(kind);
      else next.add(kind);
      return next;
    });

  const candidateChanges = candidate?.changes ?? [];
  const changeCounts = candidateChanges.reduce(
    (counts, change) => ({ ...counts, [change.change]: counts[change.change] + 1 }),
    { changed: 0, added: 0, removed: 0, unchanged: 0 },
  );
  const candidateContents = candidate?.bundle.contents ?? [];
  const reusedSize = candidateChanges
    .filter((change) => change.change === "unchanged")
    .reduce(
      (total, change) =>
        total +
        (candidateContents.find(
          (content) => content.kind === change.kind && content.name === change.name,
        )?.size ?? 0),
      0,
    );
  const candidateTotal = candidate?.bundle.total_size ?? 0;
  const visibleChanges = candidateChanges.filter(
    (change) => showUnchanged || change.change !== "unchanged",
  );
  const changeGroups = [
    { label: "Components", description: "Application workloads and configuration", kinds: ["component"] },
    { label: "Operations", description: "Actions and runbooks", kinds: ["action", "runbook"] },
    { label: "Infrastructure", description: "Sandbox and install stack", kinds: ["sandbox", "stack-asset"] },
    { label: "Runtime", description: "Runner and packaged artifacts", kinds: ["runner-image", "runner-binary", "image"] },
  ].map((group) => ({ ...group, changes: visibleChanges.filter((change) => group.kinds.includes(change.kind)) }))
    .filter((group) => group.changes.length > 0);
  const statusMatchesCandidate =
    !!candidate && status.data?.bundle_digest === candidate.bundle.bundle_digest;
  const statusSteps = statusMatchesCandidate ? (status.data?.steps ?? []) : [];
  const approvalPhase = statusMatchesCandidate
    ? (status.data?.approval_phase ?? "components")
    : "components";
  const plannedChanges = candidateChanges.filter(
    (change) =>
      (approvalPhase === "sandbox"
        ? change.kind === "sandbox"
        : change.kind === "component") &&
      (change.change === "changed" || change.change === "added"),
  );
  const plansFinished = plannedChanges.every(
    (change) =>
      change.plan_step_id &&
      candidatePlanSteps.find((step) => step.id === change.plan_step_id)?.status === "finished",
  );
  const removedComponents = candidateChanges.filter(
    (change) => change.kind === "component" && change.change === "removed",
  ).length;
  const hasRemovals = removedComponents > 0;
  const canApprove =
    !!candidate &&
    plansFinished &&
    !hasRemovals &&
    statusMatchesCandidate &&
    status.data?.approval_required === true;

  const sandboxPlanChanges = candidateChanges.filter(
    (change) => change.kind === "sandbox" && (change.change === "changed" || change.change === "added"),
  );
  const componentPlanChanges = candidateChanges.filter(
    (change) => change.kind === "component" && (change.change === "changed" || change.change === "added"),
  );

  const historyCard = (
    <section className="card bundle-history-card">
      <div className="card-header">
        <div>
          <h3>
            <Clock /> Activation history
          </h3>
          <p>Select an activation to review what changed from the bundle it replaced.</p>
        </div>
      </div>
      <State
        loading={bundle.isLoading}
        empty={!bundle.isLoading && history.length === 0}
        emptyText="No activations recorded yet"
      />
      <div className="activity">
        {history.map((info) => (
          <HistoryEntry
            key={info.bundle_digest}
            info={info}
            active={info.bundle_digest === active?.bundle_digest}
            comparison={historyComparisons.find(
              (comparison) => comparison.bundle_digest === info.bundle_digest,
            )}
            upgradeRun={runs.data?.find(
              (run) =>
                run.ref_kind === "upgrade" &&
                run.bundle_digest === info.bundle_digest,
            )}
            stackCandidate={bundle.data?.stack_candidate ?? undefined}
            expanded={expandedHistory === info.bundle_digest}
            hideDetails
            onToggle={() =>
              setExpandedHistory((digest) =>
                digest === info.bundle_digest ? null : info.bundle_digest,
              )
            }
          />
        ))}
      </div>
      {history.map((info) => {
        if (expandedHistory !== info.bundle_digest) return null;
        return (
          <HistoryEntry
            key={`detail-${info.bundle_digest}`}
            info={info}
            active={info.bundle_digest === active?.bundle_digest}
            comparison={historyComparisons.find(
              (comparison) => comparison.bundle_digest === info.bundle_digest,
            )}
            upgradeRun={runs.data?.find(
              (run) =>
                run.ref_kind === "upgrade" &&
                run.bundle_digest === info.bundle_digest,
            )}
            stackCandidate={bundle.data?.stack_candidate ?? undefined}
            expanded
            hideHeader
            onToggle={() => undefined}
          />
        );
      })}
    </section>
  );

  return (
    <main>
      <PageHeader
        title="Bundles"
        subtitle="Review uploads, plan deployments, and inspect activation history."
        actions={<button type="button" className="primary" onClick={() => setShowBundleUpload(true)}><UploadCloud /> Upload bundle</button>}
      />
      <State error={bundle.error} loading={bundle.isLoading} />
      {!bundle.isLoading && !bundle.error && !active && (
        <div className="notice" style={{ margin: "0 0 20px" }}>
          No bundle inventory published yet
          {digest ? ` (catalog reports sha256:${digest.slice(0, 16)}…)` : ""}.
          The airgapped runner publishes the inventory when it activates a
          bundle; runners older than this portal never publish it.
        </div>
      )}

      <div className="bundle-workspace-tabs" role="tablist" aria-label="Bundle workspace">
        <button type="button" className={bundleView === "review" ? "active" : ""} onClick={() => setBundleView("review")}><Package /> Review upload {candidate && <Badge theme="info">{changeCounts.changed + changeCounts.added + changeCounts.removed}</Badge>}</button>
        <button type="button" className={bundleView === "active" ? "active" : ""} onClick={() => setBundleView("active")}><ShieldCheck /> Active bundle</button>
        <button type="button" className={bundleView === "history" ? "active" : ""} onClick={() => setBundleView("history")}><Clock /> History</button>
        {active && <span className="active-bundle-summary"><Badge theme="success" mono>active</Badge><code>{shortDigest(active.bundle_digest)}</code><span>activated {formatTime(active.activated_at)}</span></span>}
      </div>

      {showBundleUpload && (
        <div className="bundle-upload-backdrop" role="presentation" onMouseDown={() => !uploadBundle.isPending && setShowBundleUpload(false)}>
          <section className="card bundle-upload-dialog" role="dialog" aria-modal="true" aria-labelledby="bundle-upload-title" onMouseDown={(event) => event.stopPropagation()}>
            <div className="card-header">
              <div><h3 id="bundle-upload-title"><UploadCloud /> Upload bundle</h3><p>Compare a bundle archive with the active installation before planning or deploying it.</p></div>
              <button type="button" className="icon-button" aria-label="Close upload" disabled={uploadBundle.isPending} onClick={() => setShowBundleUpload(false)}><XCircle /></button>
            </div>
            <label
              className={`bundle-drop-zone ${draggingBundle ? "dragging" : ""} ${uploadBundle.isPending ? "uploading" : ""}`}
              onDragEnter={(event) => { event.preventDefault(); setDraggingBundle(true); }}
              onDragOver={(event) => event.preventDefault()}
              onDragLeave={() => setDraggingBundle(false)}
              onDrop={(event) => { event.preventDefault(); setDraggingBundle(false); selectBundle(event.dataTransfer.files); }}
            >
              <input type="file" accept=".zst,.tar.zst,application/zstd" disabled={uploadBundle.isPending} onChange={(event) => selectBundle(event.target.files)} />
              {uploadBundle.isPending ? <Loader2 className="spin" /> : <UploadCloud />}
              <strong>{uploadBundle.isPending ? (bundleUploadProgress.total > 0 && bundleUploadProgress.loaded >= bundleUploadProgress.total ? "Upload complete · Processing bundle" : "Uploading bundle…") : "Drop bundle here or choose a file"}</strong>
              <span>{uploadBundle.isPending && bundleUploadProgress.total > 0 ? (bundleUploadProgress.loaded < bundleUploadProgress.total ? `${Math.min(100, Math.round((bundleUploadProgress.loaded / bundleUploadProgress.total) * 100))}% · ${formatBytes(bundleUploadProgress.loaded)} of ${formatBytes(bundleUploadProgress.total)}` : bundleUploadStatus.data?.detail || "Waiting for backend processing status…") : "Compressed .tar.zst bundle, up to 5 GiB"}</span>
            </label>
            {uploadBundle.error && <div className="notice error candidate-notice">{uploadBundle.error.message}</div>}
          </section>
        </div>
      )}

      {showClearCandidate && candidate && (
        <div className="bundle-upload-backdrop" role="presentation" onMouseDown={() => !clearCandidate.isPending && setShowClearCandidate(false)}>
          <section className="card bundle-clear-dialog" role="dialog" aria-modal="true" aria-labelledby="bundle-clear-title" onMouseDown={(event) => event.stopPropagation()}>
            <div className="card-header">
              <div><h3 id="bundle-clear-title"><XCircle /> Clear staged bundle?</h3><p>This removes it from the review queue without deleting its archive, plan run, or audit record.</p></div>
            </div>
            <div className="bundle-clear-candidate">
              <span>{candidate.archive_name || "Staged bundle"}</span>
              <code>{candidate.bundle.bundle_digest}</code>
            </div>
            {clearCandidate.error && <div className="notice error">{clearCandidate.error.message}</div>}
            <div className="bundle-clear-actions">
              <button type="button" className="secondary" disabled={clearCandidate.isPending} onClick={() => setShowClearCandidate(false)}>Cancel</button>
              <button type="button" className="secondary danger" disabled={clearCandidate.isPending} onClick={() => clearCandidate.mutate(candidate.bundle.bundle_digest)}>
                {clearCandidate.isPending ? <Loader2 className="spin" /> : <XCircle />} Clear staged bundle
              </button>
            </div>
          </section>
        </div>
      )}

      {bundleView === "review" && !candidate && !bundle.isLoading && <section className="card bundle-empty-review"><Package /><h3>No bundle awaiting review</h3><p>Upload a bundle to compare it with the active installation.</p><button type="button" className="primary" onClick={() => setShowBundleUpload(true)}><UploadCloud /> Upload bundle</button></section>}

      {bundleView === "review" && candidate && (
        <>
        <div className="review-toolbar">
          <div className="review-view-tabs" role="tablist" aria-label="Upload review">
            <button type="button" className={reviewView === "changes" ? "active" : ""} onClick={() => setReviewView("changes")}>Bundle changes <Badge>{changeCounts.changed + changeCounts.added + changeCounts.removed}</Badge></button>
            <button type="button" className={reviewView === "plan" ? "active" : ""} onClick={() => setReviewView("plan")}>Deployment plan <Badge theme={stackPlanFailed ? "error" : stackCandidate ? "info" : ""}>{stackPlanRunning ? "running" : stackPlanFailed ? "failed" : stackCandidate ? "ready" : "not run"}</Badge></button>
          </div>
          <button type="button" className="secondary danger" onClick={() => setShowClearCandidate(true)}><XCircle /> Clear staged bundle</button>
        </div>
        {reviewView === "changes" && (
        <section className="card candidate-card">
          <div className="card-header candidate-header">
            <div>
              <h3><Package /> Bundle diff</h3>
              <p className="candidate-file">
                {candidate.archive_name && <><FileArchive /><strong>{candidate.archive_name}</strong></>}
                {candidate.archive_size != null && <span>{formatBytes(candidate.archive_size)}</span>}
                <span>uploaded {formatTime(candidate.staged_at)}</span>
              </p>
            </div>
          </div>
          <div className="candidate-digests">
            <div><span>Active</span><code>{active?.bundle_digest ?? candidate.previous_digest}</code></div>
            <ChevronRight />
            <div><span>Uploaded</span><code>{candidate.bundle.bundle_digest}</code></div>
          </div>
          <div className="candidate-summary">
            {(["changed", "added", "removed", "unchanged"] as const).map((change) => (
              <div key={change}><strong>{changeCounts[change]}</strong><span>{change}</span></div>
            ))}
            <div>
              <strong>{formatBytes(reusedSize)}</strong>
              <span>content reuse of {formatBytes(candidateTotal)}</span>
            </div>
          </div>
          {hasRemovals && (
            <div className="notice error candidate-notice">
              Component removals are not supported. Remove the deleted components manually or stage a candidate without removals before deploying.
            </div>
          )}
          <div className="candidate-table-heading">
            <h3>Changes</h3>
            <button type="button" className={`chip ${showUnchanged ? "on" : ""}`} onClick={() => setShowUnchanged((value) => !value)}>
              Show unchanged ({changeCounts.unchanged})
            </button>
          </div>
          <div className="bundle-change-groups">
            {changeGroups.map((group) => (
              <details
                className="bundle-change-group"
                key={group.label}
                open={openChangeGroups.includes(group.label)}
              >
                <summary onClick={(event) => {
                  event.preventDefault();
                  setOpenChangeGroups((current) => current.includes(group.label) ? current.filter((label) => label !== group.label) : [...current, group.label]);
                }}>
                  <span><strong>{group.label}</strong><small>{group.description}</small></span>
                  <span><Badge>{group.changes.length}</Badge><ChevronRight /></span>
                </summary>
                <div className="bundle-change-list">
                {group.changes.map((change, index) => {
                  const key = `${change.kind}-${change.name}-${index}`;
                  const actionDiff = change.kind === "action" && (change.previous_action_definition || change.candidate_action_definition);
                  const componentDiff = change.kind === "component" && (change.previous_component_definition || change.candidate_component_definition);
                  const runbookDiff = change.kind === "runbook" && (change.previous_runbook_definition || change.candidate_runbook_definition);
                  const expandable = !!componentDiff || !!actionDiff || !!runbookDiff;
                  const expanded = expandedChange === key;
                  return (
                    <Fragment key={key}>
                      <button type="button" className={`bundle-change-row ${change.change === "unchanged" ? "quiet-row" : ""}`} disabled={!expandable} onClick={() => expandable && setExpandedChange(expanded ? null : key)}>
                        <span className="bundle-change-name"><ContentKindIcon kind={change.kind} /><span><strong>{change.name}</strong>{change.detail && <code>{change.detail}</code>}</span></span>
                        <BundleChangeDescription change={change} />
                        <Badge theme={change.change === "removed" ? "error" : change.change === "added" ? "success" : change.change === "changed" ? "info" : ""} mono>{change.change}</Badge>
                        {expandable && <ChevronRight className={expanded ? "expanded-chevron" : ""} />}
                      </button>
                      {expanded && (componentDiff || actionDiff || runbookDiff) && (
                        <div className="bundle-change-detail">
                          {actionDiff
                            ? <ActionDefinitionDiff before={change.previous_action_definition} after={change.candidate_action_definition} />
                            : runbookDiff
                              ? <RunbookDefinitionDiff before={change.previous_runbook_definition} after={change.candidate_runbook_definition} />
                            : componentDiff
                              ? <ComponentDefinitionDiff before={change.previous_component_definition} after={change.candidate_component_definition} />
                              : null}
                        </div>
                      )}
                    </Fragment>
                  );
                })}
                </div>
              </details>
            ))}
          </div>
        </section>
        )}
        {reviewView === "plan" && (
        <section className="card deployment-plan-card">
          <div>
            <div className="deployment-plan-heading">
              <span><Layers /> <strong>Deployment plan</strong><small>Run environment-specific plans only when you are ready.</small></span>
              <span className="deployment-plan-summary-state">
                {stackCandidate && (
                  <Badge theme={stackPlanFailed ? "error" : stackCandidate.stack_applied_at ? "success" : "info"} mono>
                    {stackPlanRunning ? "generating" : stackPlanFailed ? "failed" : stackCandidate.no_op ? "no changes" : stackCandidate.stack_applied_at ? "applied" : "ready"}
                  </Badge>
                )}
              </span>
            </div>
            <div className="deployment-plan-body">
              {stackPending && (
                <div className="stack-apply-callout">
                  <div className="stack-apply-callout-heading">
                    <span><Terminal /><strong>Next step: Apply the install stack</strong></span>
                    <Badge mono>required</Badge>
                  </div>
                  <p>The deployment plan is ready. The portal does not execute this CloudFormation change set; run the command below to apply it and activate the candidate runner.</p>
                  <div className="stack-apply-command">
                    <code>nuon-bundle upgrade apply-stack --yes</code>
                    <button type="button" className="secondary sm" onClick={async () => {
                      await navigator.clipboard.writeText("nuon-bundle upgrade apply-stack --yes");
                      setApplyCommandCopied(true);
                      window.setTimeout(() => setApplyCommandCopied(false), 2000);
                    }}>
                      {applyCommandCopied ? <CheckCircle2 /> : <Copy />} {applyCommandCopied ? "Copied" : "Copy"}
                    </button>
                  </div>
                  <small>This page will detect activation automatically and then show the bundle deployment approval.</small>
                </div>
              )}
              {candidatePlanRun && <Link className="download-link" to={`/runs?run=${encodeURIComponent(candidatePlanRun.run_id)}`}><PlayCircle /> View plan run in Runs</Link>}
              {!stackCandidate || stackPlanFailed ? (
                <button
                  type="button"
                  className="primary"
                  disabled={!candidate.deployment || planStack.isPending}
                  onClick={() => planStack.mutate(candidate.bundle.bundle_digest)}
                >
                  {planStack.isPending ? <Loader2 className="spin" /> : <Layers />}
                  {planStack.isPending ? "Starting plan…" : stackPlanFailed ? "Retry deployment plan" : "Run deployment plan"}
                </button>
              ) : stackPlanRunning ? (
                <div className="notice candidate-notice">CloudFormation is evaluating the candidate template. You can continue reviewing the bundle diff while this runs.</div>
              ) : (
                <>
                  <div className="candidate-table-heading">
                    <div><h3>Install stack plan</h3><code>{stackCandidate.change_set_name}</code></div>
                  </div>
                  {stackCandidate.no_op
                    ? <div className="notice candidate-notice">The candidate does not change any CloudFormation resources.</div>
                    : <StackChangesTable candidate={stackCandidate} heading={false} />}
                </>
              )}
              {planStack.error && <div className="notice error candidate-notice">{planStack.error.message}</div>}
              {candidatePlanRun && !stackPlanRunning && (
                <div className="candidate-environment-plans">
                  <CandidatePlanResults
                    title="Sandbox plan"
                    description="Terraform changes to shared installation infrastructure."
                    changes={sandboxPlanChanges}
                    steps={candidatePlanSteps}
                  />
                  <CandidatePlanResults
                    title="Component plans"
                    description="Helm, Terraform, and manifest changes for application components."
                    changes={componentPlanChanges}
                    steps={candidatePlanSteps}
                  />
                </div>
              )}
              {stackCandidate?.runner_activated_at && !hasRemovals && !plansFinished && (
                <div className="notice candidate-notice">{approvalPhase === "sandbox" ? "Deploy becomes available after the sandbox Terraform plan finishes." : "Deploy becomes available after all changed and added component plans finish."}</div>
              )}
              {stackCandidate?.runner_activated_at && !hasRemovals && plansFinished && (!statusMatchesCandidate || status.data?.approval_required !== true) && <div className="notice candidate-notice">Waiting for the runner to request approval.</div>}
              {stackCandidate && !stackPlanRunning && !stackPending && (
                <button type="button" className="primary" disabled={!canApprove || approve.isPending} onClick={() => approve.mutate(candidate.bundle.bundle_digest)}>
                  {approve.isPending ? <Loader2 className="spin" /> : <PlayCircle />}
                  {approvalPhase === "sandbox" ? "Approve & Deploy Sandbox" : "Approve & Deploy Bundle"}
                </button>
              )}
              {approve.error && <div className="notice error candidate-notice">{approve.error.message}</div>}
            </div>
          </div>
        </section>
        )}
        </>
      )}

      {bundleView === "history" && historyCard}

      {bundleView === "active" && active && (
        <div className="tiles">
          <div className="tile">
            <div className="tile-top">
              <span className="tile-label">Active bundle</span>
              <ShieldCheck className="good" />
            </div>
            <div className="tile-value" title={active.bundle_digest}>
              {digest?.slice(0, 12)}…
            </div>
            <div className="tile-note">
              activated {formatTime(active.activated_at)}
            </div>
          </div>
          <div className="tile">
            <div className="tile-top">
              <span className="tile-label">Target</span>
              <Cpu />
            </div>
            <div className="tile-value">
              {active.target
                ? `${active.target.os}/${active.target.architecture}`
                : "—"}
            </div>
            <div className="tile-note">bundle build platform</div>
          </div>
          <div className="tile">
            <div className="tile-top">
              <span className="tile-label">Contents</span>
              <Layers />
            </div>
            <div className="tile-value">{contents.length} items</div>
            <div className="tile-note">
              {contentSummary(contents) || "no inventory"}
            </div>
          </div>
          <div className="tile">
            <div className="tile-top">
              <span className="tile-label">Total size</span>
              <FileArchive />
            </div>
            <div className="tile-value">{formatBytes(active.total_size)}</div>
            <div className="tile-note">packaged blob bytes</div>
          </div>
        </div>
      )}

      {bundleView === "active" && <div className="grid-1-2" style={{ gridTemplateColumns: "3fr 2fr" }}>
        <div className="stack">
          {active && (
            <section className="card">
              <div className="card-header">
                <h3>
                  <FileArchive /> Contents
                </h3>
                <div className="chips">
                  {presentKinds.map(({ kind, label }) => (
                    <button
                      key={kind}
                      className={`chip ${hiddenKinds.has(kind) ? "" : "on"}`}
                      onClick={() => toggleKind(kind)}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
              <State
                empty={visible.length === 0}
                emptyText="No items match the selected kinds"
              />
              {visible.length > 0 && (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Kind</th>
                        <th>Digest</th>
                        <th style={{ textAlign: "right" }}>Size</th>
                      </tr>
                    </thead>
                    <tbody>
                      {visible.map((item, index) => (
                        <tr key={`${item.kind}-${item.name}-${index}`}>
                          <td>
                            <span className="name-cell">
                              <ContentKindIcon kind={item.kind} />
                              {item.kind === "component" ? (
                                <Link
                                  className="strong link"
                                  to={`/components/${encodeURIComponent(item.name)}`}
                                >
                                  {item.name}
                                </Link>
                              ) : (
                                <span className="strong">{item.name}</span>
                              )}
                            </span>
                            {item.detail && (
                              <code className="subtle">{item.detail}</code>
                            )}
                          </td>
                          <td>
                            <Badge theme="brand" mono>
                              {item.kind}
                            </Badge>
                          </td>
                          <td>
                            <code
                              style={{ fontSize: 11, color: "var(--muted)" }}
                              title={item.digest}
                            >
                              {item.digest
                                ?.replace("sha256:", "")
                                .slice(0, 12) ?? "—"}
                            </code>
                          </td>
                          <td
                            style={{
                              textAlign: "right",
                              fontFamily: "var(--font-mono)",
                              fontSize: 12,
                            }}
                          >
                            {formatBytes(item.size)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </section>
          )}
        </div>

        <div className="stack">
          {active && (
            <section className="card">
              <div className="card-header">
                <h3>
                  <ShieldCheck /> Verification
                </h3>
                <Badge theme="success" mono>
                  verified at activation
                </Badge>
              </div>
              <div className="verify-list">
                <VerificationRow
                  ok={active.verification.blobs_verified}
                  label="Blob digests match the bundle manifest"
                >
                  checked during extraction, before any job ran
                </VerificationRow>
                <VerificationRow
                  ok={active.verification.envelope_parsed}
                  label="Plan envelope parsed and digest-pinned"
                >
                  <span style={{ fontFamily: "var(--font-mono)" }}>
                    {active.bundle_digest}
                  </span>
                </VerificationRow>
                {active.archive_digest && (
                  <VerificationRow
                    ok
                    label="Transport archive checksum recorded (.tar.zst)"
                  >
                    <span style={{ fontFamily: "var(--font-mono)" }}>
                      {active.archive_digest}
                    </span>
                  </VerificationRow>
                )}
                <VerificationRow
                  ok={contents.length > 0}
                  label="Contents inventory published"
                >
                  {contentSummary(contents) || "empty"}
                </VerificationRow>
              </div>
            </section>
          )}

          <section className="card">
            <div className="card-header">
              <h3>
                <Package /> Installing a new bundle
              </h3>
            </div>
            <div className="card-body install-steps">
              <p>
                This portal is read-only for bundles: it shows what the
                airgapped runner activated. To upgrade:
              </p>
              <ol>
                <li>
                  Receive the new <code>.tar.zst</code> over your offline
                  transfer channel.
                </li>
                <li>
                  Verify it:{" "}
                  <code>nuon-bundle verify &lt;bundle.tar.zst&gt;</code>
                </li>
                <li>
                  Deploy it: <code>nuon-bundle deploy</code>
                </li>
              </ol>
              <p>
                Once the runner activates the new bundle, this page updates and
                the previous digest moves to history.
              </p>
            </div>
          </section>
        </div>
      </div>}
    </main>
  );
};

/* Runbooks */

const Runbooks = () => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const refs = catalog.data?.refs ?? [];
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [message, setMessage] = useState("");
  const selected = refs.find((ref) => ref.id === selectedID) ?? refs[0];
  const lastRuns = latestRunByRef(runs.data);

  const dispatch = useMutation({
    mutationFn: dispatchRef,
    onSuccess: (data) => {
      setMessage(`Dispatch ${data.dispatch_id} queued — track it under Runs.`);
      queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
    onError: (error) => setMessage(error.message),
  });

  return (
    <main>
      <PageHeader
        title="Runbooks"
        subtitle="Bundled operations: actions, runbooks, drift checks, and cron schedules."
      />
      <div className="grid-1-2" style={{ gridTemplateColumns: "3fr 2fr" }}>
        <section className="card">
          <State
            error={catalog.error}
            loading={catalog.isLoading}
            empty={!catalog.isLoading && refs.length === 0}
            emptyText="No refs in this bundle"
          />
          {refs.length > 0 && (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Kind</th>
                    <th>Last run</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {refs.map((ref) => {
                    const last = lastRuns.get(ref.id);
                    return (
                      <tr
                        key={ref.id}
                        className={`selectable ${selected?.id === ref.id ? "selected" : ""}`}
                        onClick={() => setSelectedID(ref.id)}
                      >
                        <td>
                          <span className="name-cell">
                            <KindIcon kind={ref.kind} />
                            <span className="strong">{ref.name}</span>
                          </span>
                          <code className="subtle">{ref.id}</code>
                        </td>
                        <td>
                          <Badge theme="brand" mono>
                            {ref.kind}
                          </Badge>
                        </td>
                        <td>
                          {last ? (
                            <span className="name-cell">
                              {statusTheme(last.status) === "success" ? (
                                <CheckCircle2
                                  style={{ color: "var(--success-text)" }}
                                />
                              ) : (
                                <Clock />
                              )}
                              {formatTime(last.started_at)}
                            </span>
                          ) : (
                            <span style={{ color: "var(--muted)" }}>—</span>
                          )}
                        </td>
                        <td className="actions">
                          <button
                            className="secondary sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedID(ref.id);
                            }}
                          >
                            <PlayCircle style={{ width: 13, height: 13 }} /> Run
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {selected && (
          <section className="card">
            <div className="card-header" style={{ display: "block" }}>
              <h3>
                <KindIcon kind={selected.kind} /> {selected.name}
              </h3>
              <p className="panel-desc">
                Runs inside the airgapped runner. Progress and drift verdicts
                appear under Runs.
              </p>
            </div>
            {message && (
              <div
                className={`notice ${dispatch.isError ? "error" : ""}`}
                style={{ marginTop: 16 }}
              >
                {message}
              </div>
            )}
            <div className="panel-fields">
              <div>
                <div className="field-label">details</div>
                <dl className="kv">
                  <div>
                    <dt>ref</dt>
                    <dd>{selected.id}</dd>
                  </div>
                  <div>
                    <dt>kind</dt>
                    <dd>{selected.kind}</dd>
                  </div>
                  <div>
                    <dt>component</dt>
                    <dd>{selected.component ?? "—"}</dd>
                  </div>
                  <div>
                    <dt>cron</dt>
                    <dd>{selected.cron_schedule ?? "on-demand"}</dd>
                  </div>
                  <div>
                    <dt>last run</dt>
                    <dd>{formatTime(lastRuns.get(selected.id)?.started_at)}</dd>
                  </div>
                </dl>
              </div>
            </div>
            <div className="panel-footer">
              {lastRuns.get(selected.id) && (
                <button className="secondary" onClick={() => navigate("/runs")}>
                  View history
                </button>
              )}
              <button
                className="primary"
                disabled={dispatch.isPending}
                onClick={() => dispatch.mutate(selected.id)}
              >
                <PlayCircle /> {dispatch.isPending ? "Queuing…" : "Execute"}
              </button>
            </div>
          </section>
        )}
      </div>
    </main>
  );
};

/* Runs */

const StageIcon = ({ status }: { status: string }) => {
  const theme = statusTheme(status);
  if (theme === "success") return <CheckCircle2 className="ok" />;
  if (theme === "error") return <XCircle className="fail" />;
  if (theme === "info") return <Loader2 className="run" />;
  return <Circle className="idle" />;
};

const DriftBox = ({ drift, jobID }: { drift: TDrift; jobID?: string }) => (
  <div className="drift-box">
    <div className="drift-head">
      <Badge theme={drift.drifted ? "error" : "success"} mono>
        {drift.drifted ? "drifted" : "no drift"}
      </Badge>
      {drift.summary && <span>{drift.summary}</span>}
    </div>
    <dl>
      <div>
        <dt>Resource changes</dt>
        <dd>{drift.resource_changes}</dd>
      </div>
      <div>
        <dt>Output changes</dt>
        <dd>{drift.output_changes}</dd>
      </div>
      <div>
        <dt>Resource drift</dt>
        <dd>{drift.resource_drift}</dd>
      </div>
    </dl>
    <DriftResources drift={drift} jobID={jobID} />
  </div>
);

const CloudFormationAppliedChanges = () => {
  const stack = useQuery({
    queryKey: ["install-stack"],
    queryFn: getInstallStack,
  });
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const created = Array.from(
    new Map(
      (stack.data?.events ?? [])
        .filter(
          (event) =>
            event.status === "CREATE_COMPLETE" &&
            event.logical_resource_id &&
            event.logical_resource_id !== stack.data?.name,
        )
        .map((event) => [event.logical_resource_id, event]),
    ).values(),
  );
  if (created.length === 0) return null;
  return (
    <div className="stage-changes">
      <div className="stage-changes-heading">
        <strong>Applied changes</strong>
        <span>{created.length} CloudFormation resources created</span>
      </div>
      <ul className="drift-resource-list">
        {created.map((event) => (
          <li className="drift-resource-item" key={event.logical_resource_id}>
            <button
              type="button"
              className={`drift-resource-toggle${expanded.has(event.logical_resource_id!) ? " expanded" : ""}`}
              onClick={() => setExpanded((current) => {
                const next = new Set(current);
                if (next.has(event.logical_resource_id!)) next.delete(event.logical_resource_id!);
                else next.add(event.logical_resource_id!);
                return next;
              })}
              disabled={!stack.data?.resources?.[event.logical_resource_id!]}
              title={stack.data?.resources?.[event.logical_resource_id!] ? "Show template properties" : "Template properties unavailable"}
            >
              <ChevronRight className="drift-resource-chevron" />
              <ChangeBadge action="create" />
              <code>{event.logical_resource_id}</code>
              <span>{event.resource_type}</span>
            </button>
            {expanded.has(event.logical_resource_id!) && stack.data?.resources?.[event.logical_resource_id!] && (
              <div className="drift-resource-detail">
                <PlanResourceDiff
                  resource={{
                    address: event.logical_resource_id!,
                    change: { before: null, after: stack.data.resources[event.logical_resource_id!].properties },
                  }}
                  action="create"
                  changedOnly
                />
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};

const Stage = ({ step, steps }: { step: TRunStep; steps: TRunStep[] }) => {
  const isPlan = step.name.includes("create-apply-plan");
  const applyID = step.id.endsWith("-plan")
    ? `${step.id.slice(0, -"-plan".length)}-apply`
    : "";
  const applied =
    isPlan && steps.some((candidate) => candidate.id === applyID && candidate.status === "finished");
  return (
    <div className="stage">
      <StageIcon status={step.status} />
      <div className="stage-body">
        <div className="stage-top">
          <span className="name">{step.name}</span>
          <code>{step.kind}</code>
        </div>
        <div className="stage-meta">
          <Badge>{step.status}</Badge>
          {step.status === "auto-skipped" && step.source_run_id && (
            <span>
              from{" "}
              <Link
                className="link"
                to={`/runs?run=${encodeURIComponent(step.source_run_id)}`}
              >
                successful source run
              </Link>
            </span>
          )}
          {step.status_description && <span>{step.status_description}</span>}
          {step.started_at && <span>started {formatTime(step.started_at)}</span>}
          {step.started_at && step.finished_at && (
            <span>{formatDuration(step.started_at, step.finished_at)}</span>
          )}
          {step.job_id && (
            <span>
              job{" "}
              <Link
                className="job-link"
                to={`/logs?job=${encodeURIComponent(step.job_id)}`}
                title="View job logs"
              >
                <code>{step.job_id}</code>
              </Link>
            </span>
          )}
        </div>
        {step.error && <pre className="error-box">{step.error}</pre>}
        {step.drift && <DriftBox drift={step.drift} jobID={step.job_id} />}
        {step.kind === "cloudformation" && step.status === "finished" && (
          <CloudFormationAppliedChanges />
        )}
        {isPlan && step.status === "finished" && (
          <div className="stage-changes">
            <div className="stage-changes-heading">
              <strong>{applied ? "Applied changes" : "Planned changes"}</strong>
              <span>
                {applied
                  ? `Applied by ${steps.find((candidate) => candidate.id === applyID)?.name}`
                  : "The corresponding apply stage has not completed"}
              </span>
            </div>
            <StepPlanPreview step={step} />
          </div>
        )}
      </div>
    </div>
  );
};

type TRunPhase = {
  id: string;
  label: string;
  steps: TRunStep[];
};

const runPhases = (run: TRun): TRunPhase[] => {
  const phases = new Map<string, TRunPhase>();
  const add = (id: string, label: string, step: TRunStep) => {
    const phase = phases.get(id) ?? { id, label, steps: [] };
    phase.steps.push(step);
    phases.set(id, phase);
  };
  for (const step of run.steps ?? []) {
    const identity = `${step.id} ${step.name} ${step.kind}`.toLowerCase();
    if (identity.includes("install-stack") || step.kind === "cloudformation") {
      add("stack", "Stack", step);
    } else if (identity.includes("sandbox")) {
      add("sandbox", "Sandbox", step);
    } else if (["installation", "upgrade", "bundle-plan"].includes(run.ref_kind)) {
      add("components", "Components", step);
    } else {
      add("execution", "Execution", step);
    }
  }
  return Array.from(phases.values());
};

const phaseStatus = (phase: TRunPhase) => {
  if (phase.steps.some((step) => statusTheme(step.status) === "error")) return "failed";
  if (phase.steps.some((step) => statusTheme(step.status) === "info")) return "in-progress";
  if (phase.steps.every((step) => statusTheme(step.status) === "success" || step.status === "auto-skipped")) return "finished";
  return "available";
};

const preferredPhaseStep = (phase: TRunPhase) =>
  phase.steps.find((step) => step.name.includes("create-apply-plan")) ??
  phase.steps.find((step) => step.error || step.drift) ??
  phase.steps[0];

const RunDetail = ({ runId, initial }: { runId: string; initial: TRun }) => {
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: ["run", runId],
    queryFn: () => getRun(runId),
    initialData: initial,
  });
  const run = detail.data;
  const phases = runPhases(run);
  const [selectedPhaseID, setSelectedPhaseID] = useState("");
  const [selectedStepID, setSelectedStepID] = useState("");
  const selectedPhase = phases.find((phase) => phase.id === selectedPhaseID) ?? phases[0];
  const selectedStep = selectedPhase?.steps.find((step) => step.id === selectedStepID) ??
    (selectedPhase ? preferredPhaseStep(selectedPhase) : undefined);
  const control = useMutation({
    mutationFn: (action: "retry" | "user-skip" | "cancel") => controlRun(runId, action),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["run", runId] });
      void queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
  });
  const pendingRetry = run.status === "failed-pending-retry";
  return (
    <div className="runs-detail-stack">
      <section className="card run-overview">
        <div className="card-header run-overview-header">
          <div>
            <h3>
              {run.ref_name || run.ref_id} <Badge>{run.status}</Badge>
            </h3>
            <div className="stage-meta" style={{ marginTop: 8 }}>
              <span><code>{run.run_id}</code></span>
              <span>source {run.source}</span>
              <span>started {formatTime(run.started_at)}</span>
              {run.finished_at && <span>{formatDuration(run.started_at, run.finished_at)}</span>}
              {run.bundle_digest && (
                <span title={run.bundle_digest}>
                  bundle <code>{run.bundle_digest.replace("sha256:", "").slice(0, 12)}…</code>
                </span>
              )}
              {run.result_directive && <span>directive <code>{run.result_directive}</code></span>}
            </div>
          </div>
          {run.ref_kind !== "bundle-plan" && (pendingRetry || run.status === "in-progress") && (
            <div className="actions run-controls">
              {pendingRetry && <button onClick={() => control.mutate("retry")} disabled={control.isPending}>Retry</button>}
              {pendingRetry && <button onClick={() => control.mutate("user-skip")} disabled={control.isPending}>Skip failed step</button>}
              <button onClick={() => control.mutate("cancel")} disabled={control.isPending}>Cancel</button>
            </div>
          )}
        </div>
        {run.error && <pre className="error-box run-error">{run.error}</pre>}
        {phases.length > 0 && (
          <div className="run-phase-nav">
            {phases.map((phase, index) => (
              <button
                type="button"
                className={selectedPhase?.id === phase.id ? "active" : ""}
                key={phase.id}
                onClick={() => {
                  setSelectedPhaseID(phase.id);
                  setSelectedStepID(preferredPhaseStep(phase).id);
                }}
              >
                <span className="run-phase-number">{index + 1}</span>
                <span className="run-phase-label">{phase.label}<small>{phase.steps.length} {phase.steps.length === 1 ? "step" : "steps"}</small></span>
                <StageIcon status={phaseStatus(phase)} />
              </button>
            ))}
          </div>
        )}
        {(run.events ?? []).length > 0 && (
          <details className="run-events">
            <summary>Immutable event history ({run.events?.length})</summary>
            <ol>{run.events?.map((event) => <li key={event.sequence}><code>{event.sequence}</code> {formatTime(event.created_at)} — {event.status.status}{event.status.result_directive ? ` (${event.status.result_directive})` : ""}</li>)}</ol>
          </details>
        )}
      </section>

      <section className="card run-stage-inspector">
        {!selectedPhase || !selectedStep ? (
          <div className="state">No steps recorded</div>
        ) : (
          <>
            <div className="card-header run-inspector-header">
              <div><h3>{selectedPhase.label}</h3><p>{selectedPhase.steps.length} recorded {selectedPhase.steps.length === 1 ? "step" : "steps"}</p></div>
              <div className="run-step-tabs">
                {selectedPhase.steps.map((step) => (
                  <button
                    type="button"
                    className={selectedStep.id === step.id ? "active" : ""}
                    key={step.id}
                    onClick={() => setSelectedStepID(step.id)}
                  >
                    <StageIcon status={step.status} />
                    {step.name.replace(" create-apply-plan", " plan").replace(" apply-plan", " apply")}
                  </button>
                ))}
              </div>
            </div>
            <div className="run-selected-stage">
              <Stage step={selectedStep} steps={run.steps ?? []} />
            </div>
          </>
        )}
      </section>
    </div>
  );
};

type RunFilter = "all" | "install" | "plan" | "action" | "failed";

const Runs = () => {
  const runs = useQuery({ queryKey: ["runs"], queryFn: getRuns });
  const [params, setParams] = useSearchParams();
  const [filter, setFilter] = useState<RunFilter>("all");
  const selectedID = params.get("run");
  const sorted = [...(runs.data ?? [])].sort(
    (a, b) => Date.parse(b.started_at) - Date.parse(a.started_at),
  );
  const filtered = sorted.filter((run) => {
    if (filter === "install") return ["installation", "upgrade"].includes(run.ref_kind);
    if (filter === "plan") return run.ref_kind === "bundle-plan";
    if (filter === "action") return !["installation", "upgrade", "bundle-plan"].includes(run.ref_kind);
    if (filter === "failed") return statusTheme(run.status) === "error";
    return true;
  });
  const selected = filtered.find((run) => run.run_id === selectedID) ?? filtered[0];
  const setSelectedID = (id: string) =>
    setParams({ run: id }, { replace: true });
  const runSummary = (run: TRun) => {
    const steps = run.steps ?? [];
    const skipped = steps.filter(
      (step) => step.status === "auto-skipped",
    ).length;
    const executed = steps.filter(
      (step) =>
        step.status !== "auto-skipped" &&
        (step.started_at ||
          step.finished_at ||
          !["installation", "upgrade"].includes(run.ref_kind)),
    ).length;
    return skipped > 0
      ? `${executed} executed · ${skipped} auto-skipped`
      : `${executed} executed`;
  };

  return (
    <main>
      <PageHeader
        title="Runs"
        subtitle="Installation, deployment plans, bundle upgrades, and runbook execution history."
      />
      <div className="runs-layout">
        <section className="card runs-history">
          <div className="card-header runs-history-header">
            <div><h3>Run history</h3><p>{filtered.length} of {sorted.length} runs</p></div>
            <div className="run-filters">
              {(["all", "install", "plan", "action", "failed"] as RunFilter[]).map((value) => (
                <button
                  type="button"
                  className={filter === value ? "active" : ""}
                  key={value}
                  onClick={() => setFilter(value)}
                >
                  {value === "all" ? "All" : value[0].toUpperCase() + value.slice(1)}
                </button>
              ))}
            </div>
          </div>
          <State
            error={runs.error}
            loading={runs.isLoading}
            empty={!runs.isLoading && filtered.length === 0}
            emptyText={sorted.length === 0 ? "No runs recorded yet" : "No runs match this filter"}
          />
          {filtered.length > 0 && (
            <div className="table-wrap runs-history-table">
              <table>
                <thead>
                  <tr>
                    <th>Run</th>
                    <th>Type</th>
                    <th>Started</th>
                    <th>Duration</th>
                    <th>Progress</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((run) => (
                    <tr
                      key={run.run_id}
                      className={`selectable ${selected?.run_id === run.run_id ? "selected" : ""}`}
                      onClick={() => setSelectedID(run.run_id)}
                    >
                      <td>
                        <span className="name-cell">
                          <KindIcon kind={run.ref_kind} />
                          <span className="strong">
                            {run.ref_name || run.ref_id}
                          </span>
                        </span>
                        <code className="subtle">{run.run_id}</code>
                      </td>
                      <td><Badge theme="brand" mono>{run.ref_kind}</Badge></td>
                      <td>
                        <span style={{ fontSize: 12, color: "var(--muted)" }}>
                          {formatTime(run.started_at)}
                        </span>
                      </td>
                      <td><span className="run-table-muted">{run.finished_at ? formatDuration(run.started_at, run.finished_at) : "in progress"}</span></td>
                      <td><span className="run-table-muted">{runSummary(run)}</span></td>
                      <td>
                        <Badge>{run.status}</Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
        {selected && <RunDetail runId={selected.run_id} initial={selected} />}
      </div>
    </main>
  );
};

/* Logs */

const levelClass = (level?: string) => {
  const l = level?.toLowerCase() ?? "";
  if (["error", "fatal", "panic", "dpanic"].includes(l)) return "lvl-error";
  if (l === "warn") return "lvl-warn";
  if (l === "debug") return "lvl-debug";
  return "lvl-info";
};

const formatLogTime = (value?: string) =>
  value
    ? new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      }).format(new Date(value))
    : "";

const fieldsText = (fields?: Record<string, unknown>) =>
  fields
    ? Object.entries(fields)
        .map(
          ([key, value]) =>
            `${key}=${typeof value === "string" ? value : JSON.stringify(value)}`,
        )
        .join("  ")
    : "";

const entryText = (entry: TLogEntry) =>
  `${entry.msg ?? ""} ${entry.raw ?? ""} ${fieldsText(entry.fields)}`.toLowerCase();

const sourceBadgeTheme = (source?: string) =>
  source === "day2" ? "brand" : source === "install" ? "info" : "";

const LogStream = ({ job }: { job: TJobLogSummary }) => {
  const log = useQuery({
    queryKey: ["log", job.job_id],
    queryFn: () => getJobLog(job.job_id),
    enabled: job.logs_available,
  });
  const [query, setQuery] = useState("");
  const [level, setLevel] = useState<"all" | "info" | "warn" | "error" | "debug">("all");
  const entries = log.data?.entries ?? [];
  const filtered = entries.filter((entry) => {
    const matchesQuery = !query || entryText(entry).includes(query.toLowerCase());
    const matchesLevel = level === "all" || entry.level?.toLowerCase() === level;
    return matchesQuery && matchesLevel;
  });

  return (
    <div className="logs-detail-stack">
      <section className="card log-overview">
        <div className="card-header log-overview-header">
          <div>
            <h3>
              <Terminal /> {job.name || job.job_id}{" "}
              {job.status && <Badge>{job.status}</Badge>}
              {!job.logs_available && <Badge theme="warn">logs not synced</Badge>}
            </h3>
            <div className="stage-meta log-overview-meta" style={{ marginTop: 8 }}>
              <span><code>{job.job_id}</code></span>
              {job.source && <span>source <code>{job.source}</code></span>}
              {job.ref_name && <span>ref {job.ref_name}</span>}
              {job.run_id && <span>run <code>{job.run_id}</code></span>}
              {job.started_at && <span>started {formatTime(job.started_at)}</span>}
            </div>
          </div>
          <div className="log-overview-actions">
            {job.run_id && (
              <Link className="secondary-link" to={`/runs?run=${encodeURIComponent(job.run_id)}`}>
                <ScrollText /> View run
              </Link>
            )}
            {job.logs_available && (
              <a className="secondary-link" href={jobLogDownloadURL(job.job_id)} download>
                <Download /> Download .ndjson
              </a>
            )}
          </div>
        </div>
      </section>

      <section className="card log-output">
        <div className="card-header log-output-header">
          <div>
            <h3>Log output</h3>
            <p>{log.data?.total ?? 0} recorded lines{log.data?.truncated ? ` · showing the latest ${entries.length}` : ""}</p>
          </div>
          <div className="log-level-filters">
            {(["all", "info", "warn", "error", "debug"] as const).map((value) => (
              <button type="button" className={level === value ? "active" : ""} key={value} onClick={() => setLevel(value)}>
                {value === "all" ? "All levels" : value}
              </button>
            ))}
          </div>
        </div>
        <div className="search-row log-search-row">
          <Search />
          <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Filter messages and structured fields…" />
          <span className="polling" style={{ whiteSpace: "nowrap" }}>{filtered.length}/{entries.length} lines</span>
        </div>
        <State
          error={log.error}
          loading={log.isLoading}
          empty={!log.isLoading && !log.error && filtered.length === 0}
          emptyText={!job.logs_available ? "This job ran, but its log file was not synced." : query || level !== "all" ? "No log lines match these filters" : "No log lines yet"}
        />
        {filtered.length > 0 && (
          <div className="log-stream">
            {filtered.map((entry, index) => (
              <div className="log-row" key={index}>
                <span className="log-time" title={entry.time ? formatDate(entry.time) : undefined}>{formatLogTime(entry.time)}</span>
                <span className={`log-level ${levelClass(entry.level)}`}>{entry.raw ? "—" : entry.level}</span>
                <span className="log-msg">
                  {entry.raw ?? entry.msg}
                  {entry.fields && <span className="log-fields"> {fieldsText(entry.fields)}</span>}
                </span>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
};

type LogJobFilter = "all" | "install" | "day2" | "failed";

const Logs = () => {
  const [params, setParams] = useSearchParams();
  const logs = useQuery({ queryKey: ["logs"], queryFn: getLogs });
  const [jobQuery, setJobQuery] = useState("");
  const [jobFilter, setJobFilter] = useState<LogJobFilter>("all");
  const jobs = [...(logs.data?.jobs ?? [])].sort(
    (a, b) => Date.parse(b.started_at ?? "") - Date.parse(a.started_at ?? ""),
  );
  const selectedID = params.get("job");
  const filteredJobs = jobs.filter((job) => {
    const matchesQuery = !jobQuery || `${job.job_id} ${job.name ?? ""} ${job.ref_name ?? ""} ${job.run_id ?? ""}`
      .toLowerCase()
      .includes(jobQuery.toLowerCase());
    if (!matchesQuery) return false;
    if (jobFilter === "install") return job.source === "install";
    if (jobFilter === "day2") return job.source === "day2";
    if (jobFilter === "failed") return statusTheme(job.status) === "error";
    return true;
  });
  const selected = filteredJobs.find((job) => job.job_id === selectedID) ?? filteredJobs[0];

  useEffect(() => {
    if (selected && !selectedID)
      setParams({ job: selected.job_id }, { replace: true });
  }, [selected, selectedID, setParams]);

  return (
    <main>
      <PageHeader
        title="Logs"
        subtitle="Job logs synced by the airgapped runner: install steps and day-2 runs."
      />
      <div className="logs-layout">
        <section className="card logs-history">
          <div className="card-header logs-history-header">
            <div><h3>Job history</h3><p>{filteredJobs.length} of {jobs.length} jobs</p></div>
            <div className="log-job-filters">
              {(["all", "install", "day2", "failed"] as LogJobFilter[]).map((value) => (
                <button type="button" className={jobFilter === value ? "active" : ""} key={value} onClick={() => setJobFilter(value)}>
                  {value === "all" ? "All" : value === "day2" ? "Day-2" : value[0].toUpperCase() + value.slice(1)}
                </button>
              ))}
            </div>
          </div>
          <div className="search-row log-job-search">
            <Search />
            <input
              value={jobQuery}
              onChange={(event) => setJobQuery(event.target.value)}
              placeholder="Filter by job, run, or operation…"
            />
          </div>
          <State
            error={logs.error}
            loading={logs.isLoading}
            empty={!logs.isLoading && !logs.error && filteredJobs.length === 0}
            emptyText={jobs.length === 0 ? "No job logs synced yet" : "No jobs match these filters"}
          />
          {filteredJobs.length > 0 && (
            <div className="table-wrap logs-history-table">
              <table>
                <thead>
                  <tr>
                    <th>Job</th>
                    <th>Run</th>
                    <th>Started</th>
                    <th>Source</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredJobs.map((job) => (
                    <tr
                      key={job.job_id}
                      className={`selectable ${selected?.job_id === job.job_id ? "selected" : ""}`}
                      onClick={() => setParams({ job: job.job_id })}
                    >
                      <td>
                        <span className="name-cell">
                          <Terminal />
                          <span className="strong">
                            {job.name || job.job_id}
                          </span>
                        </span>
                        <code className="subtle">{job.job_id}</code>
                        {!job.logs_available && <span className="subtle">log file not synced</span>}
                      </td>
                      <td><span className="run-table-muted">{job.ref_name ?? job.run_id ?? "—"}</span></td>
                      <td><span className="run-table-muted">{formatTime(job.started_at)}</span></td>
                      <td>
                        {job.source ? (
                          <Badge theme={sourceBadgeTheme(job.source)} mono>
                            {job.source}
                          </Badge>
                        ) : (
                          <span style={{ color: "var(--faint)" }}>—</span>
                        )}
                      </td>
                      <td>
                        {job.status ? (
                          <Badge>{job.status}</Badge>
                        ) : (
                          <span style={{ color: "var(--faint)" }}>—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
        {selected && <LogStream job={selected} key={selected.job_id} />}
      </div>
    </main>
  );
};

/* Shell */

const RunnerPill = () => {
  const runner = useRunnerState();
  if (runner.state === "unknown")
    return (
      <span className="top-pill">
        <CircleDot /> runner unknown
      </span>
    );
  return (
    <span className="top-pill">
      <CircleDot className={runner.state === "active" ? "ok" : "stale"} />{" "}
      runner {runner.state}
    </span>
  );
};

const Layout = () => {
  const catalog = useQuery({ queryKey: ["catalog"], queryFn: getCatalog });
  const digest = catalog.data?.bundle_digest;
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="mark">
            <ShieldCheck />
          </span>
          <span>
            <span className="brand-name">Airgap Portal</span>
            <span className="brand-sub">offline · airgapped runner</span>
          </span>
        </div>
        <nav>
          <NavLink end to="/">
            <LayoutDashboard /> Dashboard
          </NavLink>
          <NavLink to="/components">
            <Box /> Components
          </NavLink>
          <NavLink to="/sandbox">
            <Cpu /> Sandbox
          </NavLink>
          <NavLink to="/stack">
            <Layers /> Stack
          </NavLink>
          <NavLink to="/bundles">
            <FileArchive /> Bundles
          </NavLink>
          <NavLink to="/runbooks">
            <PlayCircle /> Runbooks
          </NavLink>
          <NavLink to="/runs">
            <ScrollText /> Runs
          </NavLink>
          <NavLink to="/logs">
            <Terminal /> Logs
          </NavLink>
        </nav>
      </aside>
      <div className="main-col">
        <header className="topbar">
          <div className="topbar-id">
            <Package />
            <span className="value">
              {catalog.data?.deployment_id ?? "loading…"}
            </span>
            <span className="sep">·</span>
            <ShieldCheck className="verified" />
            <span className="value" title={digest}>
              {digest ? digest.replace("sha256:", "").slice(0, 16) : "loading…"}
            </span>
          </div>
          <div className="top-pills">
            <span className="top-pill">
              <WifiOff className="offline" /> air-gapped
            </span>
            <RunnerPill />
          </div>
        </header>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route
            path="/status"
            element={<Navigate to="/components" replace />}
          />
          <Route path="/sandbox" element={<SandboxPage />} />
          <Route path="/stack" element={<StackPage />} />
          <Route path="/components" element={<Components />} />
          <Route path="/components/:name" element={<ComponentDetail />} />
          <Route path="/bundles" element={<Bundles />} />
          <Route path="/runbooks" element={<Runbooks />} />
          <Route path="/runs" element={<Runs />} />
          <Route path="/logs" element={<Logs />} />
          <Route path="/refs" element={<Runbooks />} />
        </Routes>
      </div>
    </div>
  );
};

export const App = () => <Layout />;
