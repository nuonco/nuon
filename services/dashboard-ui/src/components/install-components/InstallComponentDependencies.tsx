'use client'

import { usePathname } from 'next/navigation'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import {
  ComponentsTooltip,
  getContextTooltipItemsFromInstallComponents,
} from '@/components/components/ComponentsTooltip'
import { useInstall } from '@/hooks/use-install'
import type { TInstallComponent } from '@/types'

interface IInstallComponentDependencies {
  deps: string[]
}

export const InstallComponentDependencies = ({
  deps,
}: IInstallComponentDependencies) => {
  const pathname = usePathname()
  const { install } = useInstall()

  const depIds = new Set(deps?.map((d) => d) ?? [])
  const depSummaries = getContextTooltipItemsFromInstallComponents(
    install.install_components.filter((ic) =>
      depIds.has(ic.component_id)
    ) as TInstallComponent[],
    pathname
  )

  return depSummaries?.length === 0 ? (
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
