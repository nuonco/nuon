import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import type { TInstallComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export function getContextTooltipItemsFromInstallComponents(
  components: TInstallComponent[],
  pathname: string
): TContextTooltipItem[] {
  return components.map((comp) => ({
    id: comp?.component_id,
    href: `${pathname}/${comp?.component_id}`,
    leftContent: (
      <Status
        status={comp?.status_v2?.status}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
    title: comp?.component?.name,
    subtitle: toSentenceCase(comp?.status_v2?.status),
  }))
}

interface IComponentsTooltip {
  children: React.ReactNode
  componentSummaries: TContextTooltipItem[]
  title: string
  position?: 'top' | 'bottom' | 'left' | 'right'
}

export const ComponentsTooltip = ({
  children,
  componentSummaries,
  title,
  position = 'right',
  ...props
}: IComponentsTooltip) => {
  return (
    <ContextTooltip
      showCount
      items={componentSummaries}
      title={title}
      position={position}
      {...props}
    >
      {children}
    </ContextTooltip>
  )
}
