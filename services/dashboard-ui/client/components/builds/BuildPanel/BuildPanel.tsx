import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { CompositeError } from '@/components/common/CompositeError'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AppBranchRunCard } from '@/components/branches/AppBranchRunCard'
import { BuildImageSource } from '@/components/builds/BuildImageSource'
import { ComponentType } from '@/components/components/ComponentType'
import { LogsPanel } from '@/components/log-stream/LogsPanel'
import { RunSummary } from '@/components/runs/RunSummary'
import { TracePanel } from '@/components/spans/TracePanel'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'
import { useOrg } from '@/hooks/use-org'
import { getComponent } from '@/lib'
import { BuildProvider } from '@/providers/build-provider'
import type { TComponent } from '@/types'
import { isImageBuild } from '@/utils/image-ref'

export interface IBuildPanel extends Omit<IPanel, 'heading'> {
  buildId: string
  componentId: string
  title: string
}

const BuildPanelContent = ({
  component,
}: {
  component: TComponent | undefined
}) => {
  const { build } = useBuild()
  const { app } = useApp()
  const { org } = useOrg()
  const status = build?.status_v2?.status
  const activity = build?.status_v2?.status_human_description
  const jobs = build?.runner_job ? [build.runner_job] : []
  const triggeredBy = build?.created_by?.email ? (
    <Text variant="subtext">{build.created_by.email}</Text>
  ) : build?.created_by_id ? (
    <ID>{build.created_by_id}</ID>
  ) : null

  return (
    <>
      <div className="flex flex-wrap items-center gap-3">
        <Status status={status ?? 'unknown'} variant="badge" />
        <ID>{build?.id}</ID>
        <Time time={build?.created_at} format="relative" variant="subtext" />
        {activity && status !== 'error' ? (
          <Text variant="subtext" theme="neutral">
            {activity}
          </Text>
        ) : null}
      </div>
      {build?.composite_error ? (
        <CompositeError error={build.composite_error} />
      ) : null}
      {build?.status_v2?.metadata?.duplicate_build ? (
        <Banner theme="warn">
          <div className="flex flex-col">
            <Text weight="strong" variant="base">
              Duplicate build
            </Text>
            <Text theme="neutral">
              This build was triggered against the same commit and config as a
              previous build. Push new changes to your branch to create a unique
              build.
            </Text>
          </div>
        </Banner>
      ) : null}
      <Card className="!p-4 !shadow-none w-full">
        <div className="flex flex-wrap items-start gap-x-8 gap-y-4 [&>*]:shrink-0 [&>*]:whitespace-nowrap">
          <LabeledValue label="Status">
            <Status status={status ?? 'unknown'} />
          </LabeledValue>
          {component?.type ? (
            <LabeledValue label="Type">
              <ComponentType
                type={component.type}
                displayVariant="name"
                variant="subtext"
              />
            </LabeledValue>
          ) : null}
          {component?.name && org?.id && app?.id ? (
            <LabeledValue label="Component">
              <Link
                href={`/${org.id}/apps/${app.id}/components/${component.id}`}
                textVariant="subtext"
              >
                <Text as="span" variant="subtext" family="mono">
                  {component.name}
                </Text>
              </Link>
            </LabeledValue>
          ) : null}
          <LabeledValue label="Duration">
            <Duration
              variant="subtext"
              beginTime={build?.created_at}
              endTime={build?.updated_at}
            />
          </LabeledValue>
          {triggeredBy ? (
            <LabeledValue label="Triggered by">{triggeredBy}</LabeledValue>
          ) : null}
        </div>
      </Card>
      <Tabs
        naturalHeight
        initActiveTab="summary"
        tabsClassName="mt-4"
        tabControlsClassName="!gap-2 md:!gap-6 [&>button]:px-1 md:[&>button]:px-3"
        tabs={{
          summary: (
            <div className="flex flex-col gap-6">
              <AppBranchRunCard
                appId={app?.id}
                orgId={org?.id}
                buildStatus={status}
                sourceCommit={build?.vcs_connection_commit}
                run={build?.app_branch_run}
              />
              <RunSummary
                isLoading={!build}
                status={build?.status_v2}
                statusDescription={build?.status_description}
                showTiming={false}
                jobs={jobs}
                jobHref={(job) => `/${org?.id}/runner/jobs/${job?.id}`}
              >
                {build && isImageBuild(build) ? (
                  <BuildImageSource build={build} />
                ) : null}
              </RunSummary>
            </div>
          ),
          logs: <LogsPanel logStream={build?.log_stream} />,
          trace: <TracePanel logStream={build?.log_stream} />,
        }}
      />
    </>
  )
}

const BuildPanelBody = ({
  buildId,
  componentId,
  component,
  title,
}: Pick<IBuildPanel, 'buildId' | 'componentId' | 'title'> & {
  component: TComponent | undefined
}) => (
  <BuildProvider
    buildId={buildId}
    componentId={componentId}
    componentName={component?.name ?? title}
    shouldPoll
  >
    <BuildPanelContent component={component} />
  </BuildProvider>
)

export const BuildPanel = ({
  buildId,
  componentId,
  title,
  ...panelProps
}: IBuildPanel) => {
  const { org } = useOrg()
  const { data: component } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component', org?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId }),
    enabled: !!org?.id && !!componentId,
  })

  return (
    <Panel
      heading={
        <span className="flex items-center gap-2 min-w-0">
          {component?.type ? (
            <ComponentType type={component.type} displayVariant="icon-only" />
          ) : null}
          <Text variant="base" weight="strong" className="truncate">
            {title}
          </Text>
        </span>
      }
      size="3/4"
      aria-label="Build details"
      panelKey={`build-${componentId}-${buildId}`}
      {...panelProps}
    >
      <BuildPanelBody
        buildId={buildId}
        componentId={componentId}
        component={component}
        title={title}
      />
    </Panel>
  )
}
