import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallDeployments } from '@/lib'
import {
  DEFAULT_DEPLOYMENTS_FILTER,
  DeploymentsListPresenter,
  type IDeploymentFilter,
} from './DeploymentsListPresenter'

const PAGE_LIMIT = 20

const DATE_FILTER_GTE: Record<string, string> = {
  '24h': () => new Date(Date.now() - 86_400_000).toISOString(),
  '7d': () => new Date(Date.now() - 604_800_000).toISOString(),
  '30d': () => new Date(Date.now() - 2_592_000_000).toISOString(),
} as unknown as Record<string, string>

function getDateGte(date: string): string | undefined {
  const factory = (DATE_FILTER_GTE as unknown as Record<string, () => string>)[
    date
  ]
  return factory ? factory() : undefined
}

export const DeploymentsListContainer = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams, setSearchParams] = useSearchParams()

  const page = parseInt(searchParams.get('page') ?? '0', 10) || 0
  const filter: IDeploymentFilter = {
    search: searchParams.get('search') ?? '',
    status: searchParams.get('status') ?? 'all',
    type: searchParams.get('type') ?? 'all',
    component: searchParams.get('component') ?? 'all',
    date: searchParams.get('date') ?? 'all',
  }

  const { data, isLoading, error } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-deployments',
      org?.id,
      install?.id,
      page,
      filter.status,
      filter.type,
      filter.component,
      filter.date,
      filter.search,
    ],
    queryFn: () =>
      getInstallDeployments({
        orgId: org!.id,
        installId: install!.id,
        page,
        limit: PAGE_LIMIT,
        status: filter.status !== 'all' ? filter.status : undefined,
        type: filter.type !== 'all' ? filter.type : undefined,
        resource: filter.component !== 'all' ? filter.component : undefined,
        createdAtGte: getDateGte(filter.date),
        search: filter.search || undefined,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const handleFilterChange = (patch: Partial<IDeploymentFilter>) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      // Reset to page 0 when filter changes
      next.set('page', '0')
      Object.entries(patch).forEach(([key, value]) => {
        if (!value || value === 'all' || value === '') {
          next.delete(key)
        } else {
          next.set(key, value)
        }
      })
      return next
    })
  }

  const handleClearFilters = () => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      next.delete('search')
      next.delete('status')
      next.delete('type')
      next.delete('component')
      next.delete('date')
      next.set('page', '0')
      return next
    })
  }

  const handlePageChange = (newPage: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev)
      if (newPage === 0) {
        next.delete('page')
      } else {
        next.set('page', String(newPage))
      }
      return next
    })
  }

  return (
    <DeploymentsListPresenter
      deployments={data?.deployments ?? []}
      isLoading={isLoading}
      error={error as unknown as Error | null}
      page={data?.page ?? page}
      hasMore={data?.has_more ?? false}
      orgId={org?.id ?? ''}
      appId={install?.app_id ?? ''}
      installId={install?.id ?? ''}
      filter={filter}
      onFilterChange={handleFilterChange}
      onClearFilters={handleClearFilters}
      onPageChange={handlePageChange}
    />
  )
}
