import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import type { TAppBranchConfig, TInstall } from '@/types'
import { humanize } from '@/utils/string-utils'
import {
  previewDefaultsFromConfig,
} from '@/components/branches/shared/PreviewDefaultsEditor'

export interface IPreviewConfigSection {
  currentConfig?: TAppBranchConfig
  installs?: TInstall[]
  hasGithubVCS?: boolean
  isLoading?: boolean
}

export const PreviewConfigSection = ({
  currentConfig,
  installs = [],
  hasGithubVCS = false,
  isLoading = false,
}: IPreviewConfigSection) => {
  const defaults = previewDefaultsFromConfig(currentConfig?.preview_config, installs)
  const labels = Object.entries(defaults.labelSelector)
  const installName =
    installs.find((install) => install.id === defaults.installId)?.name ??
    currentConfig?.preview_config?.install_name
  const target =
    defaults.mode === 'build-only'
      ? 'Not used'
      : labels.length > 0
        ? null
        : installName ?? 'Not set'

  return (
    <Card>
      <div className="flex items-center justify-between gap-3">
        <Text variant="base" weight="strong">
          Defaults
        </Text>
        {isLoading ? (
          <Badge loading size="sm" loadingWidth={4} />
        ) : currentConfig?.config_number != null ? (
          <Badge variant="code" size="sm">
            v{currentConfig.config_number}
          </Badge>
        ) : null}
      </div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
        <LabeledValue label="Mode" loading={isLoading} loadingWidth={10}>
          <Badge size="sm">{humanize(defaults.mode)}</Badge>
        </LabeledValue>
        <LabeledValue label="Default install" loading={isLoading} loadingWidth={16}>
          {labels.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {labels.map(([key, value]) => (
                <LabelBadge
                  key={key}
                  labelKey={key}
                  labelValue={value}
                  size="sm"
                />
              ))}
            </div>
          ) : (
            <Text family={installName ? 'mono' : 'sans'} variant="subtext">
              {target}
            </Text>
          )}
        </LabeledValue>
        {hasGithubVCS ? (
          <>
            <LabeledValue label="Commit statuses" loading={isLoading} loadingWidth={8}>
              <Badge
                size="sm"
                theme={defaults.setStatuses ? 'success' : 'neutral'}
              >
                {defaults.setStatuses ? 'Enabled' : 'Disabled'}
              </Badge>
            </LabeledValue>
            <LabeledValue label="Pull request comments" loading={isLoading} loadingWidth={8}>
              <Badge
                size="sm"
                theme={defaults.comment ? 'success' : 'neutral'}
              >
                {defaults.comment ? 'Enabled' : 'Disabled'}
              </Badge>
            </LabeledValue>
          </>
        ) : null}
      </div>

      {!isLoading && !currentConfig?.preview_config ? (
        <Text variant="subtext" theme="neutral">
          Platform defaults are used until custom settings are saved.
        </Text>
      ) : null}
    </Card>
  )
}
