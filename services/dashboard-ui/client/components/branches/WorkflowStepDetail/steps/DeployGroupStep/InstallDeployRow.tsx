import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
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
          <Text variant="body" weight="strong" nowrap className="truncate">
            <Link href={installHref}>{install?.name || installId}</Link>
          </Text>
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
        <Text variant="subtext" className="shrink-0">
          <Link href={workflowHref}>View workflow</Link>
        </Text>
      )}
    </div>
  )
}
