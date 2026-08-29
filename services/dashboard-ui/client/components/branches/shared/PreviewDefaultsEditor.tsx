import { ToggleButton } from '@/components/common/ToggleButton'
import { Text } from '@/components/common/Text'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type { TAppBranchPreviewConfig, TAppBranchRunPreviewMode, TInstall } from '@/types'

export type PreviewInstallTargetMode = 'install' | 'labels'

export type IPreviewDefaults = {
  mode: TAppBranchRunPreviewMode
  installTargetMode: PreviewInstallTargetMode
  installId: string
  labelSelector: Record<string, string>
  setStatuses: boolean
  comment: boolean
}

export const defaultPreviewDefaults = (): IPreviewDefaults => ({
  mode: 'plan-only',
  installTargetMode: 'install',
  installId: '',
  labelSelector: {},
  setStatuses: true,
  comment: true,
})

export const previewDefaultsFromConfig = (
  config?: TAppBranchPreviewConfig,
  installs?: TInstall[]
): IPreviewDefaults => {
  const base = defaultPreviewDefaults()
  if (!config) return base

  const mode = config.mode ?? base.mode
  const setStatuses = config.set_statuses ?? true
  const comment = config.comment ?? true

  if (config.label_selector?.match_labels && Object.keys(config.label_selector.match_labels).length > 0) {
    return {
      mode,
      installTargetMode: 'labels',
      installId: '',
      labelSelector: config.label_selector.match_labels,
      setStatuses,
      comment,
    }
  }

  let installId = config.install_id ?? ''
  if (!installId && config.install_name && installs) {
    installId = installs.find((i) => i.name === config.install_name)?.id ?? ''
  }

  return {
    mode,
    installTargetMode: 'install',
    installId,
    labelSelector: {},
    setStatuses,
    comment,
  }
}

export const previewDefaultsToConfig = (
  defaults: IPreviewDefaults,
  installs: TInstall[]
): TAppBranchPreviewConfig => {
  const install = installs.find((i) => i.id === defaults.installId)
  const config: TAppBranchPreviewConfig = {
    mode: defaults.mode,
    set_statuses: defaults.setStatuses,
    comment: defaults.comment,
  }

  if (defaults.installTargetMode === 'labels' && Object.keys(defaults.labelSelector).length > 0) {
    config.label_selector = { match_labels: defaults.labelSelector }
    return config
  }

  if (defaults.installId) {
    config.install_id = defaults.installId
  } else if (install?.name) {
    config.install_name = install.name
  }

  return config
}

interface IPreviewDefaultsEditor {
  value: IPreviewDefaults
  onChange: (value: IPreviewDefaults) => void
  availableInstalls: TInstall[]
  hasGithubVCS: boolean
  disabled?: boolean
  showHeader?: boolean
}

export const PreviewDefaultsEditor = ({
  value,
  onChange,
  availableInstalls,
  hasGithubVCS,
  disabled,
  showHeader = true,
}: IPreviewDefaultsEditor) => {
  const installOptions = availableInstalls.map((i) => ({
    value: i.id,
    label: i.name,
  }))

  return (
    <div className="flex flex-col gap-4">
      {showHeader && (
        <div>
          <Text variant="base" weight="strong">
            Preview defaults
          </Text>
          <Text variant="subtext" theme="neutral">
            Default settings for preview runs on this branch.
          </Text>
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Text variant="subtext" weight="strong">
          Mode
        </Text>
        <ToggleButton<TAppBranchRunPreviewMode>
          value={value.mode}
          onChange={(mode) => onChange({ ...value, mode })}
          options={[
            { value: 'plan-only', label: 'Plan only' },
            { value: 'apply', label: 'Apply' },
            { value: 'build-only', label: 'Build only' },
          ]}
        />
      </div>

      {value.mode !== 'build-only' && (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Default install
          </Text>
          <Select
            options={installOptions}
            value={value.installId}
            onChange={(installId) => onChange({ ...value, installId, installTargetMode: 'install' })}
            placeholder="Select an install"
            disabled={disabled || installOptions.length === 0}
            menuPlacement="bottom"
          />
        </div>
      )}

      {hasGithubVCS && (
        <div className="flex flex-col gap-2">
          <CheckboxInput
            id="preview-set-statuses"
            checked={value.setStatuses}
            onChange={(e) => onChange({ ...value, setStatuses: e.target.checked })}
            disabled={disabled}
            labelProps={{ labelText: 'Set commit statuses' }}
          />
          <CheckboxInput
            id="preview-pr-comment"
            checked={value.comment}
            onChange={(e) => onChange({ ...value, comment: e.target.checked })}
            disabled={disabled}
            labelProps={{ labelText: 'Comment on pull request' }}
          />
        </div>
      )}
    </div>
  )
}
