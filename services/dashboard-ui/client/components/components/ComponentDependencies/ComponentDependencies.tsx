import { Badge } from '@/components/common/Badge'
import { Link } from '@/components/common/Link'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromComponents,
} from '../ComponentsTooltip'
import type { TComponent } from '@/types'

const INLINE_LIMIT = 2

interface IComponentDependencies {
  deps: string[]
  variant?: 'count' | 'inline'
  components: TComponent[]
  isLoading: boolean
  basePath: string
  pathname: string
  tooltipTitle?: string
}

export const ComponentDependencies = ({
  deps,
  variant = 'count',
  components,
  isLoading,
  basePath,
  pathname,
  tooltipTitle,
}: IComponentDependencies) => {
  if (variant === 'inline') {
    if (isLoading) {
      return <Skeleton height="24px" width="120px" />
    }

    const visible = components.slice(0, INLINE_LIMIT)
    const overflow = components.slice(INLINE_LIMIT)
    const overflowItems = getContextTooltipItemsFromComponents(
      overflow,
      basePath
    )

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
            title={tooltipTitle ?? "More dependencies"}
            componentSummaries={overflowItems}
          >
            <Badge variant="code">+{overflow.length}</Badge>
          </ComponentsTooltip>
        )}
      </div>
    )
  }

  const depSummaries = getContextTooltipItemsFromComponents(
    components,
    pathname
  )

  return isLoading ? (
    <Skeleton height="27px" width="33px" />
  ) : depSummaries?.length === 0 ? (
    <Icon variant="MinusIcon" />
  ) : (
    <ComponentsTooltip
      title={tooltipTitle ?? "Total dependencies"}
      componentSummaries={depSummaries}
    >
      <Badge variant="code">{depSummaries?.length}</Badge>
    </ComponentsTooltip>
  )
}
