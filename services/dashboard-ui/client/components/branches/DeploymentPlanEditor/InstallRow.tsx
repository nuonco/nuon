import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { SimpleInstallStatuses } from '@/components/installs/InstallStatuses'
import type { TCloudPlatform, TInstall } from '@/types'

interface IInstallRow {
  install: TInstall
  onRemove?: () => void
  disabled?: boolean
}

export const InstallRow = ({ install, onRemove, disabled }: IInstallRow) => {
  const platform = (install.cloud_platform?.toLowerCase() as TCloudPlatform) || 'unknown'
  const region = install.aws_account?.region || install.gcp_account?.region
  const location = install.azure_account?.location
  const hasRegion = platform !== 'unknown' && !!(region || location)
  const labelEntries = Object.entries(install.labels ?? {})

  return (
    <div className="flex items-center gap-3 px-3 py-2.5 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900">
      <div className="flex flex-col gap-1 min-w-0 flex-1">
        <div className="flex items-center gap-2 min-w-0 flex-wrap">
          <Text variant="body" weight="strong" nowrap className="truncate">
            {install.name || install.id}
          </Text>
          {labelEntries.map(([k, v]) => (
            <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" variant="code" />
          ))}
        </div>
        <div className="flex items-center gap-2 min-w-0">
          <ID className="text-[11px] font-mono text-cool-grey-400 dark:text-cool-grey-500 truncate">
            {install.id}
          </ID>
          {platform !== 'unknown' && (
            <CloudPlatform
              platform={platform}
              colorVariant="color"
              displayVariant="icon-only"
              iconSize="14"
              className="shrink-0"
            />
          )}
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
      </div>

      <div className="shrink-0">
        <SimpleInstallStatuses install={install} isLabelHidden />
      </div>

      {onRemove && (
        <Button
          variant="ghost"
          size="xs"
          onClick={onRemove}
          disabled={disabled}
          title={`Remove ${install.name || install.id}`}
          className="!p-1 shrink-0"
        >
          <Icon variant="XIcon" size={14} />
        </Button>
      )}
    </div>
  )
}
