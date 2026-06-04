import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TBannerTheme } from '@/components/common/Banner'
import type { TInstall } from '@/types'

export interface IDeprovisionBanner {
  install: TInstall
  orgId: string
  workflowId?: string
}

const LIFECYCLE_CONFIG: Record<
  string,
  { heading: string; theme: TBannerTheme }
> = {
  provisioning: {
    heading: 'This install is being provisioned',
    theme: 'info',
  },
  deprovisioning: {
    heading: 'This install is being deprovisioned',
    theme: 'warn',
  },
  deprovisioned: {
    heading: 'This install has been deprovisioned',
    theme: 'warn',
  },
}

export const DeprovisionBanner = ({
  install,
  orgId,
  workflowId,
}: IDeprovisionBanner) => {
  const status = install?.lifecycle_status?.status ?? ''
  const config = LIFECYCLE_CONFIG[status]
  if (!config) return null

  return (
    <Banner theme={config.theme} className="rounded-none border-x-0 border-t-0">
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col gap-0.5">
          <Text weight="strong">{config.heading}</Text>
          {install?.lifecycle_status?.status_human_description && (
            <Text variant="subtext">
              {install.lifecycle_status.status_human_description}
            </Text>
          )}
        </div>
        {workflowId && (
          <Link
            href={`/${orgId}/installs/${install.id}/workflows/${workflowId}`}
            className="flex items-center gap-1 text-sm shrink-0"
          >
            View workflow <Icon variant="ArrowRightIcon" size={14} />
          </Link>
        )}
      </div>
    </Banner>
  )
}
