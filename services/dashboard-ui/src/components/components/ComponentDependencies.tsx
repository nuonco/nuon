'use client'

import { usePathname } from 'next/navigation'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { useInstall } from '@/hooks/use-install'
import type { TComponent, TInstallComponent } from '@/types'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromInstallComponents,
} from './ComponentsTooltip'

interface IComponentDependencies {
  deps: Array<TComponent>
}

// TODO: make this for app component deps
export const ComponentDependencies = ({ deps }: IComponentDependencies) => {
  const pathname = usePathname()
  const { install } = useInstall()

  const depIds = new Set(deps?.map((d) => d.id) ?? [])
  const depSummaries = getContextTooltipItemsFromInstallComponents(
    install.install_components.filter((ic) =>
      depIds.has(ic.component_id)
    ) as TInstallComponent[],
    pathname
  )

  return depSummaries?.length === 0 ? (
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
