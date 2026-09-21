import { Badge } from '@/components/common/Badge'
import { cn } from '@/utils/classnames'
import type { TDeploymentAffectedResources } from './types'

export const DeploymentAffectedResourceBadges = ({
  affectedResources,
  className,
}: {
  affectedResources: TDeploymentAffectedResources
  className?: string
}) => (
  <div className={cn('flex flex-wrap gap-2', className)}>
    {affectedResources.stack && (
      <Badge size="sm" theme="neutral">
        Stack
      </Badge>
    )}
    {affectedResources.sandbox && (
      <Badge size="sm" theme="neutral">
        Sandbox
      </Badge>
    )}
    {affectedResources.components.map((component) => (
      <Badge key={component} size="sm" variant="code" theme="neutral">
        {component}
      </Badge>
    ))}
    {affectedResources.images.map((image) => (
      <Badge key={image} size="sm" variant="code" theme="neutral">
        {image}
      </Badge>
    ))}
  </div>
)
