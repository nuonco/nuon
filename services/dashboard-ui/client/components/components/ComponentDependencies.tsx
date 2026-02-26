import { useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromComponents,
} from './ComponentsTooltip'

interface IComponentDependencies {
  deps: string[]
}

export const ComponentDependencies = ({ deps }: IComponentDependencies) => {
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: result, isLoading } = useQuery({
    queryKey: ['components', org?.id, app?.id, 'deps', deps],
    queryFn: () => getComponents({
      orgId: org.id,
      appId: app.id,
      component_ids: deps.toString(),
    }),
    enabled: !!org?.id && !!app?.id && deps.length > 0,
  })

  const depSummaries = getContextTooltipItemsFromComponents(
    result?.data ?? [],
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
