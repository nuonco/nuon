'use client'

import { usePathname } from 'next/navigation'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useQueryParams } from '@/hooks/use-query-params'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromComponents,
} from './ComponentsTooltip'

interface IComponentDependencies {
  deps: string[]
}

export const ComponentDependencies = ({ deps }: IComponentDependencies) => {
  const pathname = usePathname()
  const { org } = useOrg()
  const { app } = useApp()
  const params = useQueryParams({ component_ids: deps.toString() })
  const { data: components, isLoading } = useQuery({
    path: `/api/orgs/${org?.id}/apps/${app?.id}/components${params}`,
  })

  const depSummaries = getContextTooltipItemsFromComponents(
    components,
    pathname
  )

  return isLoading ? (
    <Skeleton height="27px" width="33px" />
  ) : depSummaries?.length === 0 ? (
    <Icon variant="Minus" />
  ) : (
    <ComponentsTooltip
      title="Total dependencies"
      componentSummaries={depSummaries}
    >
      <Badge variant="code">{depSummaries?.length}</Badge>
    </ComponentsTooltip>
  )
}
