import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { cn } from '@/utils/classnames'

export interface IFleetFilterDropdown {
  id: string
  label: string
  count: number
  onReset: () => void
  menuClassName?: string
  children: ReactNode
}

export const FleetFilterDropdown = ({
  id,
  label,
  count,
  onReset,
  menuClassName,
  children,
}: IFleetFilterDropdown) => (
  <Dropdown
    buttonText={
      <span className="flex items-center gap-2">
        {label}
        {count ? (
          <Badge size="sm" theme="info">
            {count}
          </Badge>
        ) : null}
      </span>
    }
    icon={<Icon variant="FunnelIcon" size="14" />}
    iconAlignment="left"
    id={id}
    alignment="right"
    closeOnBlur={false}
  >
    <Menu className={cn('min-w-56 max-h-96', menuClassName)}>
      <div className="-mx-2 -mt-2 flex min-h-0 flex-col gap-0.5 overflow-y-auto overflow-x-hidden px-2 pt-2">
        {children}
      </div>
      {count ? <hr /> : null}
      {count ? (
        <Button className="shrink-0" variant="ghost" onClick={onReset}>
          Reset
        </Button>
      ) : null}
    </Menu>
  </Dropdown>
)
