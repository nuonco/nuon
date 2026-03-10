import { useLocation } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromComponents,
} from '@/components/components/ComponentsTooltip'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'

const INLINE_LIMIT = 2

interface IInstallComponentDependencies {
  deps: string[]
  variant?: 'count' | 'inline'
}

export const InstallComponentDependencies = ({
  deps,
  variant = 'count',
}: IInstallComponentDependencies) => {
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: result, isLoading } = useQuery({
    queryKey: ['components', org?.id, install?.app_id, 'deps', deps],
    queryFn: () =>
      getComponents({
        orgId: org.id,
        appId: install.app_id,
        component_ids: deps.toString(),
      }),
    enabled: !!org?.id && !!install?.app_id && deps?.length > 0,
  })

  if (variant === 'inline') {
    if (isLoading) {
      return <Skeleton height="24px" width="120px" />
    }

    const components = result?.data ?? []
    const basePath = `/${org.id}/installs/${install.id}/components`
    const visible = components.slice(0, INLINE_LIMIT)
    const overflow = components.slice(INLINE_LIMIT)
    const overflowItems = getContextTooltipItemsFromComponents(overflow, basePath)

    return (
      <div className="flex items-center gap-2 flex-wrap">
        {visible.map((comp) => (
          <Badge key={comp.id} variant="code">
            <Link href={`${basePath}/${comp.id}`}>
              {comp.name}
            </Link>
          </Badge>
        ))}
        {overflow.length > 0 && (
          <ComponentsTooltip
            title="More dependencies"
            componentSummaries={overflowItems}
          >
            <Badge variant="code">+{overflow.length}</Badge>
          </ComponentsTooltip>
        )}
      </div>
    )
  }

  const depSummaries = getContextTooltipItemsFromComponents(
    result?.data ?? [],
    pathname
  )

  return isLoading ? (
    <Skeleton height="27px" width="33px" />
  ) : depSummaries?.length === 0 ? (
    <Icon variant="MinusIcon" />
  ) : (
    <ComponentsTooltip
      title="Total dependencies"
      componentSummaries={depSummaries}
    >
      <Badge variant="code">{depSummaries?.length}</Badge>
    </ComponentsTooltip>
  )
}
