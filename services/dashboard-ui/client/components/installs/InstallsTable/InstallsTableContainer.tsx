import type { ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { useInstallHref } from '@/hooks/use-install-path'
import { useNewInstallIA } from '@/hooks/use-new-install-ia'
import { useOrg } from '@/hooks/use-org'
import {
  getAppInstalls,
  getInstalls,
  getInstallLabelKeys,
  getBranches,
  getAppLabels,
  toLabelColorMap,
} from '@/lib'
import type { TPaginatedResult } from '@/lib/api'
import type { TAPIError, TInstall } from '@/types'
import { CreateInstallButton } from '../CreateInstall'
import { InstallBranchFilter } from '../InstallBranchFilter'
import {
  INSTALL_HEALTH_FILTER_VALUES,
  INSTALL_LIFECYCLE_PHASE_FILTER_VALUES,
  InstallHealthFilters,
  InstallsTable,
  parseInstallsToTableData,
  type TInstallsTableScope,
} from './InstallsTable'

const LIMIT = 20

const parseListParam = (values: string[]) => [
  ...new Set(
    values
      .flatMap((value) => value.split(','))
      .map((v) => v.trim())
      .filter(Boolean)
  ),
]

const invalidFilterMessage = (
  param: string,
  values: string[],
  allowed: readonly string[]
) => {
  const invalid = values.filter((v) => !allowed.includes(v))
  if (invalid.length === 0) return undefined
  return `Unsupported ${param} filter value: ${invalid.join(', ')}. Clear the filter to see installs.`
}

export const InstallsTableContainer = ({
  appId,
  branchId,
  cloudConnectionId,
  emptyStateAction,
  emptyTitle,
  emptyMessage,
  pollInterval = 20000,
  shouldPoll = true,
}: {
  appId?: string
  branchId?: string
  cloudConnectionId?: string
  emptyStateAction?: ReactNode
  emptyTitle?: string
  emptyMessage?: string
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { org } = useOrg()
  const installHref = useInstallHref()
  const newInstallIA = useNewInstallIA()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') || undefined
  const labels = searchParams.get('labels') || undefined
  const branches = searchParams.get('branches') || undefined
  const healthValues = parseListParam(searchParams.getAll('health'))
  const lifecyclePhaseValues = parseListParam(
    searchParams.getAll('lifecycle_phase')
  )
  const health = healthValues.join(',') || undefined
  const lifecycle_phase = lifecyclePhaseValues.join(',') || undefined
  const filterError =
    invalidFilterMessage(
      'health',
      healthValues,
      INSTALL_HEALTH_FILTER_VALUES
    ) ??
    invalidFilterMessage(
      'lifecycle',
      lifecyclePhaseValues,
      INSTALL_LIFECYCLE_PHASE_FILTER_VALUES
    )
  const hasActiveFilters = Boolean(
    q || labels || branches || health || lifecycle_phase
  )

  const setListParam = (
    key: 'health' | 'lifecycle_phase',
    allowed: readonly string[],
    selected: Set<string>
  ) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        const values = allowed.filter((v) => selected.has(v))
        if (values.length === 0) params.delete(key)
        else params.set(key, values.join(','))
        params.delete('offset')
        return params
      },
      { replace: true }
    )
  }

  const scope: TInstallsTableScope = branchId ? 'branch' : appId ? 'app' : 'org'

  const {
    data: result,
    error,
    isLoading,
  } = useQuery<TPaginatedResult<TInstall[]>, TAPIError>({
    queryKey: [
      'installs',
      org.id,
      appId,
      branchId,
      cloudConnectionId,
      offset,
      q,
      labels,
      branches,
      health,
      lifecycle_phase,
    ],
    queryFn: () =>
      appId
        ? getAppInstalls({
            orgId: org.id,
            appId,
            offset,
            limit: LIMIT,
            q,
            app_branch_id: branchId,
            health,
            lifecycle_phase,
          })
        : getInstalls({
            orgId: org.id,
            cloud_connection_id: cloudConnectionId,
            offset,
            limit: LIMIT,
            q,
            labels,
            branches,
            health,
            lifecycle_phase,
          }),
    enabled: !filterError,
    retry: (failureCount, err) =>
      !(err?.status && err.status >= 400 && err.status < 500) &&
      failureCount < 3,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const errorMessage =
    filterError ??
    (error
      ? error?.error || 'Unable to load installs. Try refreshing the page.'
      : undefined)
  const installs = errorMessage ? [] : (result?.data ?? [])

  const healthFilters = (
    <InstallHealthFilters
      health={new Set(healthValues)}
      lifecyclePhase={new Set(lifecyclePhaseValues)}
      onHealthChange={(selected) =>
        setListParam('health', INSTALL_HEALTH_FILTER_VALUES, selected)
      }
      onLifecyclePhaseChange={(selected) =>
        setListParam(
          'lifecycle_phase',
          INSTALL_LIFECYCLE_PHASE_FILTER_VALUES,
          selected
        )
      }
    />
  )
  const appIds = [
    ...new Set(installs.map((i) => i.app_id).filter(Boolean)),
  ] as string[]

  const labelQueries = useQueries({
    queries: appIds.map((id) => ({
      queryKey: ['app-labels', org.id, id],
      queryFn: () => getAppLabels({ orgId: org.id, appId: id }),
      enabled: !!org.id && !!id,
    })),
  })

  const labelColorsByApp: Record<string, Record<string, string>> = {}
  appIds.forEach((id, i) => {
    labelColorsByApp[id] = toLabelColorMap(labelQueries[i]?.data)
  })

  return (
    <InstallsTable
      data={parseInstallsToTableData(
        installs,
        org.id,
        labelColorsByApp,
        scope === 'org',
        (install) =>
          installHref({
            installId: install.id,
            appId: install.app_id,
          })
      )}
      isLoading={isLoading}
      emptyStateAction={emptyStateAction ?? <CreateInstallButton />}
      emptyTitle={emptyTitle}
      emptyMessage={emptyMessage}
      errorMessage={errorMessage}
      hasActiveFilters={hasActiveFilters}
      filterActions={
        <div className="flex items-center gap-3">
          {scope === 'org' ? (
            <>
              <LabelFilterDropdown
                queryKey={['install-label-keys', org.id]}
                queryFn={() => getInstallLabelKeys({ orgId: org.id })}
              />
              <InstallBranchFilter
                queryKey={['org-branch-names', org.id]}
                queryFn={async () => {
                  const { data } = await getBranches({
                    orgId: org.id,
                    limit: 100,
                  })
                  return [
                    ...new Set(data.map((b) => b.name).filter(Boolean)),
                  ].sort()
                }}
              />
            </>
          ) : null}
          {healthFilters}
        </div>
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
      scope={scope}
      showHealthCoverage={healthValues.length > 0}
      statusAxes={newInstallIA && scope === 'branch'}
    />
  )
}
