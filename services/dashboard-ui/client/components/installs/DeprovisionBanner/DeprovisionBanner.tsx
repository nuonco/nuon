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
    heading: 'Provisioning install',
    theme: 'info',
  },
  deprovisioning: {
    heading: 'Deprovisioning install',
    theme: 'warn',
  },
  deprovisioned: {
    heading: 'Install deprovisioned',
    theme: 'warn',
  },
}

export const DeprovisionBanner = ({
  install,
  orgId,
  workflowId,
}: IDeprovisionBanner) => {
  const phase = install?.lifecycle_phase?.phase ?? ''
  const config = LIFECYCLE_CONFIG[phase]
  if (!config) return null

  const description = install?.lifecycle_phase?.description

  return (
    <Banner theme={config.theme}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col gap-0.5">
          <Text weight="strong">{config.heading}</Text>
          {description && (
            <Text variant="subtext">{description}</Text>
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
