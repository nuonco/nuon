import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
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
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds, getComponents, getSandboxBuilds } from '@/lib'
import type { TComponentType } from '@/types'
import { cacheBadgeTheme } from '../../shared/format'

interface IBuildStep {
  metadata: Record<string, any>
  status?: string
  appBranchRunId?: string
}

const isSandboxBuild = (build: any) =>
  build.component_type === 'sandbox' || build.component_id === 'sandbox'

interface IBuildRowDetail {
  build: any
  orgId?: string
  appId?: string
  appBranchRunId?: string
}

const BuildRowDetail = ({ build, orgId, appId, appBranchRunId }: IBuildRowDetail) => {
  const isSandbox = isSandboxBuild(build)
  const componentId = build.component_id as string | undefined

  const { data: componentBuildsResult, isLoading: loadingComponent } = useQuery({
    queryKey: ['component-builds', orgId, componentId, appBranchRunId],
    queryFn: () => getComponentBuilds({ orgId: orgId!, componentId: componentId!, limit: 20 }),
    enabled: !isSandbox && !!orgId && !!componentId,
  })

  const { data: sandboxBuildsResult, isLoading: loadingSandbox } = useQuery({
    queryKey: ['sandbox-builds', orgId, appId],
    queryFn: () => getSandboxBuilds({ orgId: orgId!, appId: appId!, limit: 20 }),
    enabled: isSandbox && !!orgId && !!appId,
  })

  const componentBuild = useMemo(() => {
    const list = componentBuildsResult?.data
    if (!list?.length) return undefined
    if (appBranchRunId) {
      const match = list.find((b) => b.app_branch_run_id === appBranchRunId)
      if (match) return match
    }
    return list[0]
  }, [componentBuildsResult, appBranchRunId])

  const sandboxBuild = sandboxBuildsResult?.data?.at(0)

  const isLoading = isSandbox ? loadingSandbox : loadingComponent
  const detail = isSandbox ? sandboxBuild : componentBuild

  const buildHref = isSandbox
    ? sandboxBuild?.id && orgId && appId
      ? `/${orgId}/apps/${appId}/sandbox/builds/${sandboxBuild.id}`
      : undefined
    : componentBuild?.id && orgId && appId && componentId
      ? `/${orgId}/apps/${appId}/components/${componentId}/builds/${componentBuild.id}`
      : undefined

  return (
    <div className="px-4 py-4 pl-[44px] border-t bg-cool-grey-50/60 dark:bg-dark-grey-800/40">
      {isLoading ? (
        <div className="flex items-center gap-2">
          <Loading size={14} className="text-cool-grey-400" />
          <Text variant="subtext" theme="neutral">Loading build details…</Text>
        </div>
      ) : detail ? (
        <div className="flex items-start justify-between gap-4">
          <div className="flex flex-wrap items-start gap-x-8 gap-y-3">
            {detail.created_at && (
              <LabeledValue label="Started">
                <Time variant="subtext" time={detail.created_at} format="relative" />
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
                {detail.status_v2?.status_human_description || detail.status_v2?.status || detail.status}
              </Text>
            </LabeledValue>
            {componentBuild?.resolved_tag && (
              <LabeledValue label="Tag">
                <Text variant="subtext" family="mono">{componentBuild.resolved_tag}</Text>
              </LabeledValue>
            )}
            {componentBuild?.source_digest && (
              <LabeledValue label="Digest">
                <ID className="text-[12px] font-mono">{componentBuild.source_digest}</ID>
              </LabeledValue>
            )}
          </div>
          {buildHref && (
            <Link href={buildHref} className="text-sm shrink-0">
              View build
              <Icon variant="ArrowRightIcon" size={14} />
            </Link>
          )}
        </div>
      ) : (
        <Text variant="subtext" theme="neutral">Build details unavailable.</Text>
      )}
    </div>
  )
}

export interface IBuildRow {
  build: any
  type?: TComponentType
  rowId: string
  orgId?: string
  appId?: string
  appBranchRunId?: string
}

export const BuildRow = ({ build, type, rowId, orgId, appId, appBranchRunId }: IBuildRow) => {
  const isSandbox = isSandboxBuild(build)

  return (
    <Expand
      id={`build-${rowId}`}
      headerClassName="px-4 py-3"
      heading={
        <div className="flex items-center gap-3 flex-1 min-w-0">
          {isSandbox ? (
            <Icon variant="TerminalWindowIcon" size={16} className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          ) : type ? (
            <ComponentType type={type} displayVariant="icon-only" colorVariant="color" iconSize="16" />
          ) : null}

          <Text variant="body" weight="strong" nowrap className="truncate">
            {build.component_name || build.component_id}
          </Text>

          {build.cache_status && (
            <Badge theme={cacheBadgeTheme(build.cache_status)} size="sm">
              {build.cache_status}
            </Badge>
          )}

          <div className="flex-1" />

          <Status status={build.status || 'pending'} variant="badge" className="shrink-0" />
        </div>
      }
    >
      <BuildRowDetail build={build} orgId={orgId} appId={appId} appBranchRunId={appBranchRunId} />
    </Expand>
  )
}

export const BuildStep = ({ metadata, status, appBranchRunId }: IBuildStep) => {
  const { org } = useOrg()
  const { app } = useApp()
  const builds = (metadata.builds as any[]) || []

  const { data: componentsResult } = useQuery({
    queryKey: ['components', org?.id, app?.id],
    queryFn: () => getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })

  const typeMap = useMemo(() => {
    const map: Record<string, TComponentType> = {}
    for (const c of componentsResult?.data || []) {
      if (c.id && c.type) map[c.id] = c.type
    }
    return map
  }, [componentsResult])

  if (builds.length === 0) {
    return (
      <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-800 rounded-lg border">
        <Text variant="subtext" theme="neutral">
          {status === 'in-progress' ? 'Starting component builds...' : 'Waiting to start builds...'}
        </Text>
      </div>
    )
  }

  const succeededCount = builds.filter((b: any) => b.status === 'success' || b.status === 'skipped').length
  const totalDuration = builds.reduce((acc: number, b: any) => acc + (b.duration || 0), 0)

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Text variant="body" theme="neutral">
            <span className="font-semibold">{builds.length}</span> components built
          </Text>
          <span className="text-[12px] text-cool-grey-400">·</span>
          <Text variant="body" weight="strong" theme="success">
            {succeededCount} succeeded
          </Text>
        </div>
        {totalDuration > 0 && (
          <Text variant="subtext" family="mono" theme="neutral">
            {totalDuration.toFixed(1)}s total
          </Text>
        )}
      </div>

      <div className="border rounded-[10px] divide-y overflow-hidden">
        {builds.map((build: any, i: number) => {
          const rowId = String(build.component_id || i)
          return (
            <BuildRow
              key={rowId}
              build={build}
              type={build.component_type || typeMap[build.component_id]}
              rowId={rowId}
              orgId={org?.id}
              appId={app?.id}
              appBranchRunId={appBranchRunId}
            />
          )
        })}
      </div>
    </div>
  )
}
