import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TCloudPlatform, TInstall } from '@/types'

export interface IInstallDeployRow {
  installId: string
  install?: TInstall
  deployStatus?: string
  workflowHref?: string
  installHref?: string
}

export const InstallDeployRow = ({ installId, install, deployStatus, workflowHref, installHref }: IInstallDeployRow) => {
  const platform = (install?.cloud_platform?.toLowerCase() as TCloudPlatform | undefined) || 'unknown'
  const region = install?.aws_account?.region || install?.gcp_account?.region
  const location = install?.azure_account?.location
  const hasRegion = platform !== 'unknown' && !!(region || location)

  return (
    <div className="flex items-center gap-3 px-4 py-3">
      <div className="flex items-center gap-5 min-w-0">
        {installHref ? (
          <Link href={installHref} className="text-sm font-strong truncate">
            {install?.name || installId}
          </Link>
        ) : (
          <Text variant="body" weight="strong" nowrap className="block truncate">
            {install?.name || installId}
          </Text>
        )}

        <CloudPlatform
          platform={platform}
          colorVariant="color"
          displayVariant="icon-only"
          iconSize="20"
          className="shrink-0"
        />

        {hasRegion && (
          <CloudRegion
            variant="subtext"
            theme="neutral"
            nowrap
            className="shrink-0"
            platform={platform as 'aws' | 'azure' | 'gcp'}
            region={region}
            location={location}
          />
        )}
      </div>

      <div className="flex-1" />

      <Status status={deployStatus || 'pending'} variant="badge" className="shrink-0" />

      {workflowHref && (
        <Link href={workflowHref} className="text-sm shrink-0">
          View workflow
          <Icon variant="ArrowRightIcon" size={14} />
        </Link>
      )}
    </div>
  )
}
