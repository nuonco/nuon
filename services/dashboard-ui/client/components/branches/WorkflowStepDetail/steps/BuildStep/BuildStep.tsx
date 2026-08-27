import { useMemo, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { CompositeError } from '@/components/common/CompositeError'
import { Duration } from '@/components/common/Duration'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import {
  BuildTypeFilter,
  SANDBOX_FILTER,
  uniqueBuildFilterTypes,
  useBuildTypeFilter,
  type TBuildTypeFilterKey,
} from '@/components/branches/BuildTypeFilter'
import { StepStatePlaceholder } from '../../shared/StepStatePlaceholder'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getBranchRunBuilds,
  getComponentBuild,
  getComponents,
  getSandboxBuild,
} from '@/lib'
import type {
  TAppSandboxBuild,
  TBuild,
  TComponentType,
} from '@/types'
import { changeReasonBadgeTheme, changeReasonLabel } from '../../shared/format'

interface IBuildStep {
  metadata: Record<string, any>
  status?: string
  appBranchId?: string
  appBranchRunId?: string
}

const isSandboxBuild = (build: any) =>
  build.component_type === 'sandbox' || build.component_id === 'sandbox'

const buildFilterType = (
  build: any,
  typeMap: Record<string, TComponentType>
): TBuildTypeFilterKey | undefined => {
  if (isSandboxBuild(build)) return SANDBOX_FILTER
  return (build.component_type || typeMap[build.component_id]) as
    | TComponentType
    | undefined
}

const pollingBuildStatuses = new Set([
  'queued',
  'pending',
  'planning',
  'building',
  'in-progress',
])

const shouldPollBuildDetail = (detail?: TBuild | TAppSandboxBuild) => {
  const status = detail?.status_v2?.status || detail?.status
  return !detail?.composite_error && pollingBuildStatuses.has(status || '')
}

interface IBuildRowDetail {
  detail?: TBuild | TAppSandboxBuild
  buildHref?: string
  isLoading?: boolean
}

export const BuildRowDetail = ({
  detail,
  buildHref,
  isLoading = false,
}: IBuildRowDetail) => {
  return (
    <div className="px-4 py-4 pl-[44px] border-t bg-cool-grey-50/60 dark:bg-dark-grey-800/40">
      {isLoading ? (
        <div className="flex items-center gap-2">
          <Loading size={14} className="text-cool-grey-400" />
          <Text variant="subtext" theme="neutral">
            Loading build details…
          </Text>
        </div>
      ) : detail ? (
        <div className="flex flex-col gap-4">
          {detail.composite_error ? (
            <CompositeError error={detail.composite_error} />
          ) : null}
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-wrap items-start gap-x-8 gap-y-3">
              {detail.created_at && (
                <LabeledValue label="Started">
                  <Time
                    variant="subtext"
                    time={detail.created_at}
                    format="relative"
                  />
                </LabeledValue>
              )}
              {detail.created_at && (
                <LabeledValue label="Duration">
                  <Duration
                    variant="subtext"
                    beginTime={detail.created_at}
                    endTime={detail.updated_at}
                  />
                </LabeledValue>
              )}
              <LabeledValue label="Status">
                <Text variant="subtext">
                  {detail.status_v2?.status_human_description ||
                    detail.status_v2?.status ||
                    detail.status}
                </Text>
              </LabeledValue>
              {'resolved_tag' in detail && detail.resolved_tag && (
                <LabeledValue label="Tag">
                  <Text variant="subtext" family="mono">
                    {detail.resolved_tag}
                  </Text>
                </LabeledValue>
              )}
              {'source_digest' in detail && detail.source_digest && (
                <LabeledValue label="Digest">
                  <ID className="text-[12px] font-mono">
                    {detail.source_digest}
                  </ID>
                </LabeledValue>
              )}
            </div>
            {buildHref && (
              <Link href={buildHref} className="shrink-0">View build</Link>
            )}
          </div>
        </div>
      ) : (
        <Text variant="subtext" theme="neutral">
          Build details unavailable.
        </Text>
      )}
    </div>
  )
}

interface IBuildRowDetailContainer {
  build: any
  componentBuildId?: string
  sandboxBuildId?: string
  orgId?: string
  appId?: string
  isExpanded: boolean
  isLoadingBuilds?: boolean
}

const BuildRowDetailContainer = ({
  build,
  componentBuildId,
  sandboxBuildId,
  orgId,
  appId,
  isExpanded,
  isLoadingBuilds = false,
}: IBuildRowDetailContainer) => {
  const isSandbox = isSandboxBuild(build)
  const componentId = build.component_id as string | undefined

  const componentBuildQuery = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['build', orgId, componentId, componentBuildId],
    queryFn: () =>
      getComponentBuild({
        orgId: orgId!,
        componentId: componentId!,
        buildId: componentBuildId!,
      }),
    enabled:
      isExpanded &&
      !isSandbox &&
      !!orgId &&
      !!componentId &&
      !!componentBuildId,
    refetchInterval: (query) =>
      isExpanded && shouldPollBuildDetail(query.state.data) ? 5000 : false,
  })

  const sandboxBuildQuery = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['sandbox-build', orgId, appId, sandboxBuildId],
    queryFn: () =>
      getSandboxBuild({
        orgId: orgId!,
        appId: appId!,
        buildId: sandboxBuildId!,
      }),
    enabled: isExpanded && isSandbox && !!orgId && !!appId && !!sandboxBuildId,
    refetchInterval: (query) =>
      isExpanded && shouldPollBuildDetail(query.state.data) ? 5000 : false,
  })

  const detail = isSandbox ? sandboxBuildQuery.data : componentBuildQuery.data
  const isLoading =
    isExpanded &&
    (isSandbox
      ? !!sandboxBuildId && sandboxBuildQuery.isLoading
      : isLoadingBuilds ||
        (!!componentBuildId && componentBuildQuery.isLoading))
  const buildHref = isSandbox
    ? sandboxBuildId && orgId && appId
      ? `/${orgId}/apps/${appId}/sandbox/builds/${sandboxBuildId}`
      : undefined
    : componentBuildId && orgId && appId && componentId
      ? `/${orgId}/apps/${appId}/components/${componentId}/builds/${componentBuildId}`
      : undefined

  return (
    <BuildRowDetail
      detail={detail}
      buildHref={buildHref}
      isLoading={isLoading}
    />
  )
}

export interface IBuildRow {
  build: any
  type?: TComponentType
  rowId: string
  orgId?: string
  appId?: string
  componentBuildId?: string
  sandboxBuildId?: string
  isLoadingBuilds?: boolean
}

export const BuildRow = ({
  build,
  type,
  rowId,
  orgId,
  appId,
  componentBuildId,
  sandboxBuildId,
  isLoadingBuilds,
}: IBuildRow) => {
  const isSandbox = isSandboxBuild(build)
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <Expand
      id={`build-${rowId}`}
      headerClassName="px-4 py-3"
      onExpandedChange={setIsExpanded}
      heading={
        <div className="flex items-center gap-3 flex-1 min-w-0">
          {isSandbox ? (
            <Icon
              variant="TerminalWindowIcon"
              size={16}
              className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0"
            />
          ) : type ? (
            <ComponentType
              type={type}
              displayVariant="icon-only"
              colorVariant="color"
              iconSize="16"
            />
          ) : null}

          <Text variant="body" weight="strong" nowrap className="truncate">
            {build.component_name || build.component_id}
          </Text>

          {(build.change_reason || build.cache_status) && (
            <Badge
              theme={changeReasonBadgeTheme(
                build.change_reason ||
                  (build.cache_status === 'cache hit' ? 'no_changes' : undefined)
              )}
              size="sm"
            >
              {build.change_reason
                ? changeReasonLabel(build.change_reason)
                : build.cache_status}
            </Badge>
          )}

          <div className="flex-1" />

          <Status
            status={build.status || 'pending'}
            variant="badge"
            className="shrink-0"
          />
        </div>
      }
    >
      <BuildRowDetailContainer
        build={build}
        componentBuildId={componentBuildId}
        sandboxBuildId={sandboxBuildId}
        orgId={orgId}
        appId={appId}
        isExpanded={isExpanded}
        isLoadingBuilds={isLoadingBuilds}
      />
    </Expand>
  )
}

export const BuildStep = ({
  metadata,
  status,
  appBranchId,
  appBranchRunId,
}: IBuildStep) => {
  const { org } = useOrg()
  const { app } = useApp()
  const builds = (metadata.builds as any[]) || []
  const sandboxBuildId = metadata.sandbox_build_id as string | undefined

  const { data: branchBuilds, isLoading: isLoadingBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'branch-run-builds',
      org?.id,
      app?.id,
      appBranchId,
      appBranchRunId,
    ],
    queryFn: () =>
      getBranchRunBuilds({
        orgId: org!.id,
        appId: app!.id,
        branchId: appBranchId!,
        runId: appBranchRunId!,
      }),
    enabled: !!org?.id && !!app?.id && !!appBranchId && !!appBranchRunId,
    refetchInterval: status === 'in-progress' ? 5000 : false,
  })

  const { data: componentsResult } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['components', org?.id, app?.id],
    queryFn: () =>
      getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })

  const typeMap = useMemo(() => {
    const map: Record<string, TComponentType> = {}
    for (const c of componentsResult?.data || []) {
      if (c.id && c.type) map[c.id] = c.type
    }
    return map
  }, [componentsResult])

  const componentBuildMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const build of branchBuilds || []) {
      if (build.component_id && build.id) map.set(build.component_id, build.id)
    }
    return map
  }, [branchBuilds])

  const filterTypes = useMemo(
    () => uniqueBuildFilterTypes(builds.map((b: any) => buildFilterType(b, typeMap))),
    [builds, typeMap]
  )
  const filter = useBuildTypeFilter(filterTypes)

  const visibleBuilds = useMemo(
    () =>
      builds.filter((b: any) => filter.matches(buildFilterType(b, typeMap))),
    [builds, filter.matches, filter.deselected, typeMap]
  )

  const buildSummary = useMemo(() => {
    const counts = {
      source_changed: 0,
      config_changed: 0,
      no_changes: 0,
      built: 0,
    }
    for (const b of builds) {
      const reason =
        b.change_reason ||
        (b.skipped || b.status === 'skipped' ? 'no_changes' : 'source_changed')
      if (reason === 'no_changes') {
        counts.no_changes++
      } else if (reason === 'config_changed') {
        counts.config_changed++
        counts.built++
      } else if (reason === 'source_changed') {
        counts.source_changed++
        counts.built++
      }
    }
    return counts
  }, [builds])

  if (builds.length === 0) {
    return status === 'in-progress' ? (
      <StepStatePlaceholder variant="loading">
        Starting component builds
      </StepStatePlaceholder>
    ) : (
      <StepStatePlaceholder variant="pending">
        Waiting to start component builds
      </StepStatePlaceholder>
    )
  }

  const summaryParts: string[] = []
  if (buildSummary.built > 0) {
    summaryParts.push(
      `${buildSummary.built} ${buildSummary.built === 1 ? 'component' : 'components'} built`
    )
  }
  if (buildSummary.no_changes > 0) {
    summaryParts.push(
      `${buildSummary.no_changes} no changes`
    )
  }

  const totalDuration = builds.reduce(
    (acc: number, b: any) => acc + (b.duration || 0),
    0
  )

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Text variant="body" theme="neutral">
            {summaryParts.length > 0 ? summaryParts.join(' · ') : `${builds.length} components`}
          </Text>
        </div>
        {totalDuration > 0 && (
          <Text variant="subtext" family="mono" theme="neutral">
            {totalDuration.toFixed(1)}s total
          </Text>
        )}
      </div>

      {filterTypes.length > 1 && (
        <BuildTypeFilter
          types={filter.types}
          deselected={filter.deselected}
          onToggle={filter.toggle}
        />
      )}

      <div className="border rounded-[10px] divide-y overflow-hidden">
        {visibleBuilds.length === 0 ? (
          <div className="px-4 py-3">
            <Text variant="subtext" theme="neutral">
              No builds match filters
            </Text>
          </div>
        ) : (
          visibleBuilds.map((build: any, i: number) => {
            const rowId = String(build.component_id || i)
            return (
              <BuildRow
                key={rowId}
                build={build}
                type={build.component_type || typeMap[build.component_id]}
                rowId={rowId}
                orgId={org?.id}
                appId={app?.id}
                componentBuildId={
                  build.build_id || componentBuildMap.get(build.component_id)
                }
                sandboxBuildId={sandboxBuildId}
                isLoadingBuilds={isLoadingBuilds}
              />
            )
          })
        )}
      </div>
    </div>
  )
}
