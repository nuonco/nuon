import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { MODULE_READINESS_LABELS, type IModule } from '../../utils/modules'
import { Badge } from '../atoms/Badge'
import { Card } from '../atoms/Card'
import { Icon } from '../atoms/Icon'
import { Switch } from '../atoms/Switch'
import { Text } from '../atoms/Text'

export interface IModuleCard
  extends Omit<HTMLAttributes<HTMLDivElement>, 'onToggle'> {
  module: IModule
  enabled: boolean
  pinned?: boolean
  registered?: boolean
  saving?: boolean
  loading?: boolean
  onToggle?: (enabled: boolean) => void
}

export const ModuleCard = ({
  module,
  enabled,
  pinned = false,
  registered = true,
  saving = false,
  loading = false,
  onToggle,
  className,
  ...props
}: IModuleCard) => {
  const locked = pinned || !registered || !onToggle

  return (
    <Card
      padding="sm"
      className={cn('flex items-start gap-3', className)}
      {...props}
    >
      <span
        aria-hidden
        className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-surface-02 text-secondary"
      >
        <Icon variant={module.icon} size={18} />
      </span>
      <span className="flex min-w-0 flex-1 flex-col gap-1">
        <span className="flex min-w-0 flex-wrap items-center gap-2">
          <Text weight="medium" color="primary">
            {module.name}
          </Text>
          {module.readiness !== 'built' ? (
            <Badge>{MODULE_READINESS_LABELS[module.readiness]}</Badge>
          ) : null}
          {pinned ? <Badge tone="accent">Pinned by deployment</Badge> : null}
          {!registered ? <Badge>Flag missing on this API</Badge> : null}
        </span>
        <Text variant="caption" color="secondary">
          {module.description}
        </Text>
      </span>
      <Switch
        aria-label={`${module.name} module`}
        checked={enabled}
        onChange={(next) => onToggle?.(next)}
        disabled={locked}
        loading={loading || saving}
        className="shrink-0"
      />
    </Card>
  )
}
