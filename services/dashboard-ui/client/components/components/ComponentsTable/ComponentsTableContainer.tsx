import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { LabelFilterDropdown } from '@/components/common/LabelFilterDropdown'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { ManagementDropdownContainer as ManagementDropdown } from '@/components/components/management/ManagementDropdown'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponents, getComponentLabelKeys } from '@/lib'
import { ComponentsTable, parseComponentToTableData } from './ComponentsTable'

const LIMIT = 20

export const ComponentsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'components',
      org.id,
      app.id,
      offset,
      searchParams.get('q'),
      searchParams.get('types'),
      searchParams.get('labels'),
    ],
    queryFn: () =>
      getComponents({
        orgId: org.id,
        appId: app.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
        types: searchParams.get('types') || undefined,
        labels: searchParams.get('labels') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <ComponentsTable
      data={parseComponentToTableData(result?.data ?? [], org.id, app.id)}
      isLoading={isLoading}
      filterActions={
        <div className="flex items-center gap-3">
          <LabelFilterDropdown
            queryKey={['component-label-keys', org.id, app.id]}
            queryFn={() => getComponentLabelKeys({ orgId: org.id, appId: app.id })}
          />
          <ComponentTypeFilterDropdown />
          <ManagementDropdown />
        </div>
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
