import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Status } from '@/components/common/Status'
import type { TComponent, TInstallComponent } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export function getContextTooltipItemsFromComponents(
  components: TComponent[],
  pathname: string
): TContextTooltipItem[] {
  return components?.map((comp) => ({
    id: comp?.id,
    href: `${pathname}/${comp?.id}`,
    leftContent: (
      <Status
        status={comp?.status}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
    title: comp?.name,
    subtitle: toSentenceCase(comp?.status),
  }))
}

const DEPROVISIONING_COMPONENT_SUBTITLES: Record<string, string> = {
  executing: 'Tearing down',
  active: 'Waiting to teardown',
  pending: 'Tearing down',
  success: 'Teardown complete',
  error: 'Teardown failed',
}

const DEPROVISIONED_COMPONENT_SUBTITLES: Record<string, string> = {
  executing: 'Torn down',
  active: 'Torn down',
  pending: 'Torn down',
  success: 'Torn down',
  inactive: 'Torn down',
  error: 'Teardown failed',
}

function getComponentSubtitle(
  status: string | undefined,
  lifecycleStatus?: string
): string {
  if (!status) return toSentenceCase(status)
  if (lifecycleStatus === 'deprovisioned') {
    return DEPROVISIONED_COMPONENT_SUBTITLES[status] ?? toSentenceCase(status)
  }
  if (lifecycleStatus === 'deprovisioning') {
    return DEPROVISIONING_COMPONENT_SUBTITLES[status] ?? toSentenceCase(status)
  }
  return toSentenceCase(status)
}

export function getContextTooltipItemsFromInstallComponents(
  components: TInstallComponent[],
  pathname: string,
  lifecycleStatus?: string
): TContextTooltipItem[] {
  const isDeprovisioning = lifecycleStatus === 'deprovisioning'
  const isDeprovisioned = lifecycleStatus === 'deprovisioned'
  const STALE_STATUSES = ['active', 'pending', 'executing', 'queued', 'planning', 'syncing']
  const effectiveStatus = (status: string | undefined) => {
    if (!status) return status
    if (isDeprovisioned && STALE_STATUSES.includes(status)) return 'deprovisioned'
    if (isDeprovisioning && STALE_STATUSES.includes(status)) return 'deprovisioning'
    return status
  }

  return components?.map((comp) => ({
    id: comp?.component_id,
    href: `${pathname}/${comp?.component_id}`,
    leftContent: (
      <Status
        status={effectiveStatus(comp?.status_v2?.status)}
        isWithoutText
        variant="timeline"
        iconSize={16}
      />
    ),
    title: comp?.component?.name,
    subtitle: getComponentSubtitle(comp?.status_v2?.status, lifecycleStatus),
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
      items={componentSummaries || []}
      title={title}
      position={position}
      {...props}
    >
      {children}
    </ContextTooltip>
  )
}
