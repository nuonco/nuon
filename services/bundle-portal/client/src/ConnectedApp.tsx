import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Route, Routes, useLocation, useParams } from "react-router";
import { Badge } from "@/components/common/Badge";
import { BackLink } from "@/components/common/BackLink";
import { Banner } from "@/components/common/Banner";
import { Button } from "@/components/common/Button";
import { Code } from "@/components/common/Code";
import { EmptyState } from "@/components/common/EmptyState";
import { Icon } from "@/components/common/Icon";
import { ID } from "@/components/common/ID";
import { Link } from "@/components/common/Link";
import { Loading } from "@/components/common/Loading";
import { Status } from "@/components/common/Status";
import { Table } from "@/components/common/Table";
import { Tabs } from "@/components/common/Tabs";
import { Text } from "@/components/common/Text";
import { Time } from "@/components/common/Time";
import { Timeline } from "@/components/common/Timeline";
import { SSELogs } from "@/components/log-stream/SSELogs/SSELogs";
import { PageSection } from "@/components/layout/PageSection";
import { SectionHeader } from "@/components/layout/SectionHeader";
import { Panel } from "@/components/surfaces/Panel";
import { WorkflowTimelineItem } from "@/components/workflows/WorkflowTimeline/WorkflowTimelineItem";
import { WorkflowSteps as WorkflowStepsComponent } from "@/components/workflows/WorkflowSteps/WorkflowSteps";
import { StepTitle } from "@/components/workflows/step-details/StepTitle";
import { WorkflowDetailsSection } from "@/components/workflows/workflow-details/WorkflowDetailsSection/WorkflowDetailsSection";
import { WorkflowHeader } from "@/components/workflows/workflow-details/WorkflowHeader/WorkflowHeader";
import { WorkflowMetrics } from "@/components/workflows/workflow-details/WorkflowMetrics/WorkflowMetrics";
import { WorkflowStatusSection } from "@/components/workflows/workflow-details/WorkflowStatusSection/WorkflowStatusSection";
import {
  ReleaseFiles,
  releaseFileEntryCanPreview,
} from "@/components/apps/bundles/ReleaseFiles/ReleaseFiles";
import { releaseFileEntries } from "@/components/apps/bundles/ReleaseFiles/ReleaseFilesContainer";
import { useLogFilters } from "@/hooks/use-log-filters";
import { useSurfaces } from "@/hooks/use-surfaces";
import { useWorkflowMetrics } from "@/hooks/use-workflow-metrics";
import type { IPanel } from "@/components/surfaces/Panel";
import type {
  TAppReleaseWithFiles,
  TReleasePackage,
} from "@/types/ctl-api.types";
import {
  getConnectedApprovalContents,
  getConnectedRelease,
  getConnectedReleaseFileContent,
  getConnectedReleasePackage,
  getConnectedReleases,
  getConnectedReleaseUpdates,
  getConnectedWorkflow,
  getConnectedLogStreamLogs,
  getConnectedWorkflows,
  retryConnectedWorkflowStep,
  respondToConnectedApproval,
  deployConnectedRelease,
} from "./api";
import { usePortalBranding } from "./branding";
import { ConnectedApprovalPlan } from "./connected-plan";
import type {
  TConnectedRelease,
  TConnectedReleaseUpdate,
  TConnectedLog,
  TConnectedWorkflow,
  TConnectedWorkflowStep,
} from "./types";

const workflowStatus = (value?: { status?: string }) =>
  value?.status ?? "pending";

const releaseUpdatePending = (update?: TConnectedReleaseUpdate) =>
  update?.status?.metadata?.awaiting_retry === true ||
  ![
    "success",
    "error",
    "cancelled",
    "discarded",
    "user-skipped",
    "auto-skipped",
  ].includes(update?.status?.status ?? "pending");

const ErrorBanner = ({ error }: { error: Error | null }) =>
  error ? <Banner theme="error">{error.message}</Banner> : null;

const Releases = () => {
  const releases = useQuery({
    queryKey: ["connected-releases"],
    queryFn: getConnectedReleases,
    refetchInterval: 5000,
  });
  const updates = useQuery({
    queryKey: ["connected-release-updates"],
    queryFn: getConnectedReleaseUpdates,
    refetchInterval: 5000,
  });

  const latestUpdateByRelease = useMemo(() => {
    const result = new Map<string, TConnectedReleaseUpdate>();
    for (const update of updates.data ?? []) {
      if (update.app_release_id && !result.has(update.app_release_id)) {
        result.set(update.app_release_id, update);
      }
    }
    return result;
  }, [updates.data]);
  const columns = useMemo<ColumnDef<TConnectedRelease>[]>(
    () => [
      {
        accessorKey: "release.id",
        header: "Release",
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Link href={`/releases/${row.original.release.id}`}>
              <Code variant="inline" className="!px-2 !py-1 w-fit">
                {row.original.release.id}
              </Code>
            </Link>
            <Text
              family="mono"
              variant="label"
              theme="neutral"
              className="max-w-96 truncate"
            >
              {row.original.release.semantic_digest}
            </Text>
          </div>
        ),
      },
      {
        accessorKey: "active",
        header: "Status",
        cell: ({ row }) => (
          <Status
            status={
              row.original.active
                ? "success"
                : (latestUpdateByRelease.get(row.original.release.id!)?.status
                    ?.status ?? row.original.release.status)
            }
            variant="badge"
          >
            {row.original.active
              ? "Active"
              : releaseUpdatePending(
                    latestUpdateByRelease.get(row.original.release.id!),
                  )
                ? "Proposed"
                : (latestUpdateByRelease.get(row.original.release.id!)?.status
                    ?.status ?? row.original.release.status)}
          </Status>
        ),
      },
      {
        id: "contents",
        header: "Contents",
        cell: ({ row }) => (
          <Text variant="subtext">
            {row.original.release.members?.length ?? 0} items
          </Text>
        ),
      },
      {
        id: "platforms",
        header: "Platforms",
        cell: ({ row }) =>
          row.original.release.packages?.length ? (
            <div className="flex flex-col gap-1">
              {[...row.original.release.packages]
                .sort((a, b) =>
                  (a.target_platform ?? "").localeCompare(
                    b.target_platform ?? "",
                  ),
                )
                .map((pkg) => (
                  <div key={pkg.id} className="flex items-center gap-2">
                    <Status
                      status={pkg.status ?? "unknown"}
                      variant="timeline"
                      isWithoutText
                      iconSize={12}
                      title={pkg.status}
                    />
                    <Text variant="subtext">{pkg.target_platform}</Text>
                  </div>
                ))}
            </div>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          ),
      },
      {
        accessorKey: "release.created_at",
        header: "Published",
        cell: ({ row }) => (
          <Time
            variant="subtext"
            format="relative"
            time={row.original.release.created_at}
          />
        ),
      },
    ],
    [latestUpdateByRelease],
  );

  return (
    <main className="flex flex-1 flex-col overflow-y-auto">
      <PageSection>
        <SectionHeader
          variant="page"
          title="Vendor releases"
          description="Releases published by your vendor and available through the control plane."
        />
        <ErrorBanner error={releases.error ?? updates.error} />
        <Table
          columns={columns}
          data={releases.data ?? []}
          enableSearch={false}
          isLoading={releases.isLoading}
          emptyStateProps={{
            variant: "table",
            emptyTitle: "No releases available",
            emptyMessage:
              "Your vendor has not published a release for this install yet.",
          }}
        />
      </PageSection>
    </main>
  );
};

const ConnectedReleaseFiles = ({
  previousRelease,
  release,
}: {
  previousRelease?: TAppReleaseWithFiles;
  release: TAppReleaseWithFiles;
}) => {
  const packages = release.packages ?? [];
  const defaultPackageId =
    packages.find(({ status }) => status === "active")?.id ?? packages[0]?.id;
  const [selectedPackageId, setSelectedPackageId] = useState(defaultPackageId);
  const [selectedPath, setSelectedPath] = useState<string>();

  useEffect(() => {
    if (
      !selectedPackageId ||
      !packages.some(({ id }) => id === selectedPackageId)
    ) {
      setSelectedPackageId(defaultPackageId);
    }
  }, [defaultPackageId, packages, selectedPackageId]);

  const packageDetails = useQuery({
    queryKey: ["connected-release-package", selectedPackageId],
    queryFn: () => getConnectedReleasePackage(selectedPackageId!),
    enabled: Boolean(selectedPackageId),
  });
  const platform = packages.find(
    ({ id }) => id === selectedPackageId,
  )?.target_platform;
  const previousPackageId = previousRelease?.packages?.find(
    ({ target_platform }) => target_platform === platform,
  )?.id;
  const previousPackageDetails = useQuery({
    queryKey: ["connected-release-package", previousPackageId],
    queryFn: () => getConnectedReleasePackage(previousPackageId!),
    enabled: Boolean(previousPackageId),
  });
  const entries = useMemo(
    () =>
      releaseFileEntries(
        release,
        previousRelease,
        packageDetails.data?.status === "active"
          ? packageDetails.data
          : undefined,
        packageDetails.data?.status === "active" &&
          previousPackageDetails.data?.status === "active"
          ? previousPackageDetails.data
          : undefined,
      ),
    [
      packageDetails.data,
      previousPackageDetails.data,
      previousRelease,
      release,
    ],
  );

  useEffect(() => {
    if (!selectedPath || !entries.some(({ path }) => path === selectedPath)) {
      setSelectedPath(
        entries.find(({ change }) => change !== "unchanged")?.path ??
          entries[0]?.path,
      );
    }
  }, [entries, selectedPath]);

  const selected = entries.find(({ path }) => path === selectedPath);
  const currentPath =
    selected?.category === "source" &&
    selected.current &&
    releaseFileEntryCanPreview(selected)
      ? selected.path
      : undefined;
  const previousPath =
    selected?.category === "source" &&
    selected.previous &&
    releaseFileEntryCanPreview(selected)
      ? selected.path
      : undefined;
  const currentContent = useQuery({
    queryKey: ["connected-release-file", release.id, currentPath],
    queryFn: () => getConnectedReleaseFileContent(release.id!, currentPath!),
    enabled: Boolean(currentPath),
  });
  const previousContent = useQuery({
    queryKey: ["connected-release-file", previousRelease?.id, previousPath],
    queryFn: () =>
      getConnectedReleaseFileContent(previousRelease!.id!, previousPath!),
    enabled: Boolean(previousRelease?.id && previousPath),
  });

  return (
    <ReleaseFiles
      currentContent={currentContent.data}
      entries={entries}
      isContentLoading={currentContent.isFetching || previousContent.isFetching}
      onPackageChange={setSelectedPackageId}
      onSelect={setSelectedPath}
      packageOptions={packages.map((pkg) => ({
        id: pkg.id!,
        platform: pkg.target_platform ?? "Unknown platform",
        status: pkg.status,
      }))}
      packageStatus={packageDetails.data?.status}
      previousContent={previousContent.data}
      selectedPackageId={selectedPackageId}
      selectedPath={selectedPath}
    />
  );
};

const ReleasePackages = ({ packages }: { packages: TReleasePackage[] }) => (
  <div className="mt-6 flex flex-col gap-3">
    {packages.map((pkg) => (
      <div
        key={pkg.id}
        className="flex items-center justify-between gap-4 rounded-md border p-4"
      >
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <ID>{pkg.id}</ID>
            <Badge variant="code">{pkg.target_platform}</Badge>
            <Badge>{pkg.format}</Badge>
          </div>
          <Text variant="subtext" theme="neutral" family="mono">
            {pkg.archive_checksum || pkg.package_digest}
          </Text>
        </div>
        <Status status={pkg.status ?? "unknown"}>
          {pkg.status ?? "unknown"}
        </Status>
      </div>
    ))}
  </div>
);

const ReleaseDeployment = ({ update }: { update: TConnectedReleaseUpdate }) => {
  const { addPanel } = useSurfaces();
  const queryClient = useQueryClient();
  const workflow = useQuery({
    queryKey: ["connected-workflow", update.workflow_id],
    queryFn: () => getConnectedWorkflow(update.workflow_id!),
    enabled: Boolean(update.workflow_id),
    refetchInterval: 3000,
  });
  const metrics = useWorkflowMetrics(workflow.data);
  const approvalSteps =
    workflow.data?.steps?.filter((step) => step.approval) ?? [];
  const isPlanOnly = workflow.data?.plan_only;
  const workflowCompleted =
    workflow.data?.status?.status === "success" ||
    workflow.data?.status?.status === "error";
  const openStepDetails = (step: TConnectedWorkflowStep) => {
    if (!update.workflow_id) return;
    addPanel(
      <ConnectedWorkflowStepDetails
        panelKey={step.id}
        workflowId={update.workflow_id}
        initStep={step}
      />,
      step.id,
    );
  };

  const deploy = useMutation({
    mutationFn: () => deployConnectedRelease(update.app_release_id!),
    onSuccess: () =>
      void queryClient.invalidateQueries({
        queryKey: ["connected-release-updates"],
      }),
  });

  return (
    <div className="mt-6 flex flex-col gap-6">
      <div className="border rounded-md bg-white dark:bg-dark-grey-900">
        <div className="flex flex-wrap items-center justify-between gap-4 p-4">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Deployment proposal</Text>
            <Text variant="subtext" theme="neutral">
              This release remains inactive until its deployment completes
              successfully.
            </Text>
          </div>
          <div className="flex items-center gap-3">
            <Status
              status={
                workflow.data?.status?.status ??
                update.status?.status ??
                "pending"
              }
            >
              {workflow.data?.status?.status ??
                update.status?.status ??
                "pending"}
            </Status>
            {update.workflow_id ? <ID>{update.workflow_id}</ID> : null}
          </div>
        </div>
      </div>
      <ErrorBanner error={workflow.error} />
      {workflow.isLoading ? (
        <div className="flex justify-center py-12">
          <Loading />
        </div>
      ) : workflow.data ? (
        <>
          {approvalSteps.length ? (
            <div className="flex flex-col gap-3">
              <SectionHeader
                title="Deployment plans"
                description="Review generated plans before allowing this release to be applied."
              />
              {approvalSteps.map((step) =>
                step.approval ? (
                  <div key={step.approval.id} className="border rounded-md p-4 bg-white dark:bg-dark-grey-900">
                    <Text weight="strong">{step.name}</Text>
                    <Approval
                      workflowId={workflow.data.id}
                      step={step}
                      approvalId={step.approval.id}
                      responded={Boolean(step.approval.response)}
                    />
                  </div>
                ) : null,
              )}
            </div>
          ) : null}
          {isPlanOnly && workflowCompleted && workflow.data?.status?.status === "success" ? (
            <div className="border rounded-md p-4 bg-white dark:bg-dark-grey-900">
              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-1">
                  <Text weight="strong">Approve deployment</Text>
                  <Text variant="subtext" theme="neutral">
                    The infrastructure plan has been generated. Approve to
                    deploy this release to your install.
                  </Text>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="primary"
                    disabled={deploy.isPending}
                    onClick={() => deploy.mutate()}
                  >
                    {deploy.isPending ? "Deploying…" : "Approve & Deploy"}
                  </Button>
                </div>
                <ErrorBanner error={deploy.error} />
              </div>
            </div>
          ) : null}
          <div className="flex flex-col gap-3">
            <SectionHeader title="Deployment steps" />
            <WorkflowStepsComponent
              workflowSteps={metrics.workflowSteps}
              approvalPrompt={workflow.data.approval_option === "prompt"}
              planOnly={workflow.data.plan_only}
              readOnly
              onViewDetails={(step) =>
                openStepDetails(step as TConnectedWorkflowStep)
              }
              eagerStepsLoaded
              allStepsLoaded
            />
          </div>
        </>
      ) : (
        <EmptyState
          emptyTitle="Workflow is being prepared"
          emptyMessage="Deployment details will appear when the workflow is ready."
        />
      )}
    </div>
  );
};

const ReleaseDetail = () => {
  const { id = "" } = useParams();
  const releases = useQuery({
    queryKey: ["connected-releases"],
    queryFn: getConnectedReleases,
    refetchInterval: 5000,
  });
  const updates = useQuery({
    queryKey: ["connected-release-updates"],
    queryFn: getConnectedReleaseUpdates,
    refetchInterval: 5000,
  });
  const release = useQuery({
    queryKey: ["connected-release", id],
    queryFn: () => getConnectedRelease(id),
    enabled: Boolean(id),
  });
  const releaseIndex =
    releases.data?.findIndex(({ release: item }) => item.id === id) ?? -1;
  const activeReleaseId = releases.data?.find(({ active }) => active)?.release
    .id;
  const previousReleaseId =
    activeReleaseId && activeReleaseId !== id
      ? activeReleaseId
      : releaseIndex >= 0
        ? releases.data?.[releaseIndex + 1]?.release.id
        : undefined;
  const previousRelease = useQuery({
    queryKey: ["connected-release", previousReleaseId],
    queryFn: () => getConnectedRelease(previousReleaseId!),
    enabled: Boolean(previousReleaseId),
  });
  const active = releases.data?.find(
    ({ release: item }) => item.id === id,
  )?.active;
  const update = updates.data?.find((item) => item.app_release_id === id);
  const changedFiles =
    release.data && previousRelease.data
      ? releaseFileEntries(
          release.data,
          previousRelease.data,
          undefined,
          undefined,
        ).filter(({ change }) => change !== "unchanged").length
      : 0;
  const releaseTabs: Record<string, React.ReactNode> = release.data
    ? {
        files:
          releases.isLoading || previousRelease.isLoading ? (
            <div className="flex justify-center py-12">
              <Loading />
            </div>
          ) : (
            <ConnectedReleaseFiles
              release={release.data}
              previousRelease={previousRelease.data}
            />
          ),
        deployment: update ? (
          <ReleaseDeployment update={update} />
        ) : (
          <EmptyState
            emptyTitle="No deployment proposed"
            emptyMessage="Your vendor has not proposed this release for this install."
          />
        ),
      }
    : {};
  if (release.data?.packages?.length) {
    releaseTabs.packages = <ReleasePackages packages={release.data.packages} />;
  }

  return (
    <main className="flex flex-1 flex-col overflow-y-auto">
      <PageSection>
        <BackLink />
        <ErrorBanner
          error={
            release.error ??
            releases.error ??
            updates.error ??
            previousRelease.error
          }
        />
        {release.data ? (
          <>
            <SectionHeader
              variant="page"
              title="Immutable release"
              description={release.data.id}
              actions={
                active ? (
                  <Badge theme="success">Active</Badge>
                ) : update ? (
                  <Badge>
                    {releaseUpdatePending(update)
                      ? "Proposed"
                      : (update.status?.status ?? "Not active")}
                  </Badge>
                ) : undefined
              }
            />
            <div className="flex flex-wrap gap-8">
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Status
                </Text>
                <Status status={release.data.status ?? "unknown"}>
                  {release.data.status}
                </Status>
              </div>
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Published
                </Text>
                <Time time={release.data.created_at!} />
              </div>
              <div className="flex min-w-0 flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Semantic digest
                </Text>
                <ID>{release.data.semantic_digest}</ID>
              </div>
            </div>
            <Tabs
              tabLabels={{
                files: (
                  <span className="flex items-center gap-2">
                    Files
                    {changedFiles ? (
                      <Badge theme="warn">{changedFiles} changed</Badge>
                    ) : null}
                  </span>
                ),
                deployment: (
                  <span className="flex items-center gap-2">
                    Deployment
                    {update?.status?.status === "approval-awaiting" ? (
                      <Badge theme="warn">Approval required</Badge>
                    ) : null}
                  </span>
                ),
              }}
              tabs={releaseTabs}
            />
          </>
        ) : release.isLoading ? (
          <div className="flex min-h-48 items-center justify-center">
            <Loading variant="large" />
          </div>
        ) : null}
      </PageSection>
    </main>
  );
};

const Workflows = () => {
  const workflows = useQuery({
    queryKey: ["connected-workflows"],
    queryFn: getConnectedWorkflows,
    refetchInterval: 5000,
  });
  return (
    <main className="flex flex-1 flex-col overflow-y-auto">
      <PageSection>
        <SectionHeader
          variant="page"
          title="Workflows"
          description="Control-plane operations running for this install."
        />
        <ErrorBanner error={workflows.error} />
        {workflows.isLoading ? (
          <div className="flex min-h-48 items-center justify-center">
            <Loading variant="large" />
          </div>
        ) : workflows.data?.length ? (
          <Timeline<TConnectedWorkflow>
            events={workflows.data}
            getEventKey={(workflow) => workflow.id}
            pagination={{
              hasNext: false,
              offset: 0,
              limit: workflows.data.length,
            }}
            renderEvent={(workflow) => (
              <WorkflowTimelineItem
                id={workflow.id}
                title={workflow.name || workflow.type}
                href={`/workflows/${workflow.id}`}
                status={workflowStatus(workflow.status)}
                createdAt={workflow.started_at || workflow.created_at}
                finishedAt={workflow.finished_at}
                finished={Boolean(workflow.finished_at)}
              />
            )}
          />
        ) : (
          <EmptyState
            variant="table"
            emptyTitle="No workflows found"
            emptyMessage="Operations for this install will appear here."
          />
        )}
      </PageSection>
    </main>
  );
};

const Approval = ({
  workflowId,
  step,
  approvalId,
  responded,
}: {
  workflowId: string;
  step: TConnectedWorkflowStep;
  approvalId: string;
  responded: boolean;
}) => {
  const stepId = step.id;
  const queryClient = useQueryClient();
  const contents = useQuery({
    queryKey: ["connected-approval", approvalId],
    queryFn: () => getConnectedApprovalContents(workflowId, stepId, approvalId),
  });
  const respond = useMutation({
    mutationFn: (responseType: "approve" | "deny") =>
      respondToConnectedApproval(workflowId, stepId, approvalId, responseType),
    onSuccess: () =>
      void queryClient.invalidateQueries({
        queryKey: ["connected-workflow", workflowId],
      }),
  });
  return (
    <div className="flex flex-col gap-4 pt-4">
      <ErrorBanner error={contents.error} />
      {contents.isLoading ? (
        <div className="flex justify-center py-10">
          <Loading variant="large" />
        </div>
      ) : (
        <ConnectedApprovalPlan step={step} plan={contents.data} />
      )}
      {!responded && !respond.isSuccess ? (
        <div className="flex gap-2">
          <Button
            variant="primary"
            disabled={respond.isPending}
            onClick={() => respond.mutate("approve")}
          >
            Approve
          </Button>
          <Button
            variant="danger"
            disabled={respond.isPending}
            onClick={() => respond.mutate("deny")}
          >
            Deny
          </Button>
        </div>
      ) : null}
      <ErrorBanner error={respond.error} />
    </div>
  );
};

const severityNumber = (severity?: string) => {
  switch (severity?.toLowerCase()) {
    case "trace":
      return 1;
    case "debug":
      return 5;
    case "warn":
    case "warning":
      return 13;
    case "error":
      return 17;
    case "fatal":
      return 21;
    default:
      return 9;
  }
};

const normalizeSeverity = (severity?: string) => {
  const value = severity?.toLowerCase() || "info";
  return `${value.charAt(0).toUpperCase()}${value.slice(1)}`;
};

const StepLogs = ({
  step,
}: {
  step: TConnectedWorkflowStep;
}) => {
  const logStreamId = step.log_stream?.id;
  const logs = useQuery({
    queryKey: ["connected-workflow-step-logs", logStreamId],
    queryFn: () => getConnectedLogStreamLogs(logStreamId!),
    enabled: Boolean(logStreamId),
    refetchInterval: step.status?.status === "in-progress" ? 3000 : false,
  });
  const records = useMemo<TConnectedLog[]>(
    () =>
      (logs.data ?? []).map((entry) => ({
        ...entry,
        severity_text: normalizeSeverity(entry.severity_text),
        severity_number:
          entry.severity_number ?? severityNumber(entry.severity_text),
      })),
    [logs.data],
  );
  const filters = useLogFilters(records);
  const [activeLogId, setActiveLogId] = useState<string>();
  const activeLog = records.find(({ id }) => id === activeLogId);

  return (
    <div className="pt-4">
      <ErrorBanner error={logs.error} />
      <SSELogs
        filteredLogs={filters.filteredLogs ?? undefined}
        filters={filters}
        activeLog={activeLog}
        handleActiveLog={setActiveLogId}
        isLoading={logs.isLoading}
        isConnected={step.status?.status === "in-progress"}
        showDownload={false}
      />
    </div>
  );
};

const ConnectedWorkflowStepDetails = ({
  workflowId,
  initStep,
  ...props
}: {
  workflowId: string;
  initStep: TConnectedWorkflowStep;
} & IPanel) => {
  const queryClient = useQueryClient();
  const workflow = useQuery({
    queryKey: ["connected-workflow", workflowId],
    queryFn: () => getConnectedWorkflow(workflowId),
    refetchInterval: 3000,
  });
  const step =
    workflow.data?.steps?.find(({ id }) => id === initStep.id) ?? initStep;
  const retry = useMutation({
    mutationFn: () => retryConnectedWorkflowStep(workflowId, step.id),
    onSuccess: () =>
      void queryClient.invalidateQueries({
        queryKey: ["connected-workflow", workflowId],
      }),
  });
  const tabs: Record<string, React.ReactNode> = {};

  if (step.approval) {
    tabs.plan = (
      <Approval
        workflowId={workflowId}
        step={step}
        approvalId={step.approval.id}
        responded={Boolean(step.approval.response)}
      />
    );
  }
  if (step.log_stream?.id) {
    tabs.logs = <StepLogs step={step} />;
  }

  return (
    <Panel
      className="@container"
      heading={<StepTitle step={step} />}
      size="3/4"
      {...props}
    >
      <div className="flex flex-col gap-6">
        {step.status?.status === "error" ? (
          <div className="flex justify-end">
            <Button
              variant="primary"
              disabled={retry.isPending}
              onClick={() => retry.mutate()}
            >
              {retry.isPending ? "Retrying…" : "Retry step"}
            </Button>
          </div>
        ) : null}
        <ErrorBanner error={workflow.error ?? retry.error} />
        {Object.keys(tabs).length ? (
          <Tabs tabs={tabs} />
        ) : (
          <EmptyState
            variant="history"
            emptyTitle="No plan or logs available"
            emptyMessage="Details will appear when this step generates a plan or starts streaming logs."
          />
        )}
      </div>
    </Panel>
  );
};

const WorkflowDetail = () => {
  const { id = "" } = useParams();
  const { addPanel } = useSurfaces();
  const workflow = useQuery({
    queryKey: ["connected-workflow", id],
    queryFn: () => getConnectedWorkflow(id),
    refetchInterval: 3000,
  });
  const metrics = useWorkflowMetrics(workflow.data);
  const openStepDetails = (step: TConnectedWorkflowStep) =>
    addPanel(
      <ConnectedWorkflowStepDetails
        panelKey={step.id}
        workflowId={id}
        initStep={step}
      />,
      step.id,
    );

  return (
    <main className="flex flex-1 flex-col overflow-y-auto">
      <PageSection className="!gap-2">
        <ErrorBanner error={workflow.error} />
        {workflow.data ? (
          <>
            <div className="flex flex-col gap-2">
              <WorkflowHeader workflow={workflow.data} readOnly />
              <WorkflowMetrics workflow={workflow.data} {...metrics} />
              <WorkflowStatusSection workflow={workflow.data} />
              <WorkflowDetailsSection workflow={workflow.data} orgId="" />
            </div>
            <div className="mt-6 flex flex-col gap-6">
              <SectionHeader title="Workflow steps" />
              <WorkflowStepsComponent
                workflowSteps={metrics.workflowSteps}
                approvalPrompt={workflow.data.approval_option === "prompt"}
                planOnly={workflow.data.plan_only}
                readOnly
                onViewDetails={(step) =>
                  openStepDetails(step as TConnectedWorkflowStep)
                }
                eagerStepsLoaded
                allStepsLoaded
              />
            </div>
          </>
        ) : workflow.isLoading ? (
          <div className="flex min-h-48 items-center justify-center">
            <Loading variant="large" />
          </div>
        ) : null}
      </PageSection>
    </main>
  );
};

const PortalShell = () => {
  const branding = usePortalBranding();
  const location = useLocation();
  const brand = branding.data;
  return (
    <div className="flex min-h-screen bg-white text-dark-grey-900 dark:bg-dark-grey-900 dark:text-white">
      <aside className="flex w-64 shrink-0 flex-col border-r bg-cool-grey-50 p-4 dark:bg-dark-grey-800">
        <div className="mb-8 flex min-h-10 items-center gap-3 px-2">
          {brand?.logo_url ? (
            <img
              className="h-8 max-w-36 object-contain object-left"
              src={brand.logo_url}
              alt={brand.name}
            />
          ) : (
            <span className="flex h-8 w-8 items-center justify-center rounded-md bg-primary-600 text-white">
              <Icon variant="ShieldCheckIcon" size={20} />
            </span>
          )}
          <Text weight="strong" className="truncate">
            {brand?.name ?? "Deployment portal"}
          </Text>
        </div>
        <nav className="flex flex-col gap-1">
          <Link
            variant="nav"
            isActive={location.pathname.startsWith("/releases")}
            href="/releases"
          >
            <Icon variant="ArchiveIcon" /> Releases
          </Link>
          <Link
            variant="nav"
            isActive={location.pathname.startsWith("/workflows")}
            href="/workflows"
          >
            <Icon variant="FlowArrowIcon" /> Workflows
          </Link>
        </nav>
        {brand?.support_url ? (
          <div className="mt-auto border-t pt-4">
            <Link href={brand.support_url} isExternal>
              <Icon variant="ChatCircleIcon" /> Get support
            </Link>
          </div>
        ) : null}
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center justify-between border-b px-6">
          <Text flex weight="strong">
            <Icon variant="CloudCheckIcon" className="text-green-600" />{" "}
            Connected install
          </Text>
          <Badge theme="success">Connected</Badge>
        </header>
        <Routes>
          <Route path="*" element={<Releases />} />
          <Route path="/releases" element={<Releases />} />
          <Route path="/releases/:id" element={<ReleaseDetail />} />
          <Route path="/workflows" element={<Workflows />} />
          <Route path="/workflows/:id" element={<WorkflowDetail />} />
        </Routes>
      </div>
    </div>
  );
};

export const ConnectedApp = PortalShell;
