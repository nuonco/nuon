import { ToggleButton } from '@/components/common/ToggleButton'
import { Text } from '@/components/common/Text'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import type {
  TAppBranchPreviewConfig,
  TAppBranchRunPreviewMode,
  TInstall,
} from '@/types'
import { previewModeDisplayLabel } from './preview-mode'

export type PreviewInstallTargetMode = 'install' | 'labels'

export type IPreviewDefaults = {
  mode: TAppBranchRunPreviewMode
  installTargetMode: PreviewInstallTargetMode
  installId: string
  labelSelector: Record<string, string>
  setStatuses: boolean
  comment: boolean
  ignoreDrafts: boolean
}

export const defaultPreviewDefaults = (): IPreviewDefaults => ({
  mode: 'plan-only',
  installTargetMode: 'install',
  installId: '',
  labelSelector: {},
  setStatuses: true,
  comment: true,
  ignoreDrafts: true,
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
  const ignoreDrafts = config.ignore_drafts ?? true

  if (
    config.label_selector?.match_labels &&
    Object.keys(config.label_selector.match_labels).length > 0
  ) {
    return {
      mode,
      installTargetMode: 'labels',
      installId: '',
      labelSelector: config.label_selector.match_labels,
      setStatuses,
      comment,
      ignoreDrafts,
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
    ignoreDrafts,
  }
}

export const previewDefaultsToConfig = (
  defaults: IPreviewDefaults,
  installs: TInstall[]
): TAppBranchPreviewConfig => {
  if (defaults.mode === 'none') {
    return { mode: 'none' }
  }

  const install = installs.find((i) => i.id === defaults.installId)
  const config: TAppBranchPreviewConfig = {
    mode: defaults.mode,
    set_statuses: defaults.setStatuses,
    comment: defaults.comment,
    ignore_drafts: defaults.ignoreDrafts,
  }

  if (
    defaults.installTargetMode === 'labels' &&
    Object.keys(defaults.labelSelector).length > 0
  ) {
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

      <CheckboxInput
        id="preview-enabled"
        checked={value.mode !== 'none'}
        onChange={(e) =>
          onChange({
            ...value,
            mode: e.target.checked ? 'plan-only' : 'none',
          })
        }
        disabled={disabled}
        labelProps={{
          labelText: (
            <>
              <Text weight="strong">Enable automated previews</Text>
              <Text variant="subtext" theme="neutral">
                Turn off to skip automated preview runs. Manual previews still
                work.
              </Text>
            </>
          ),
          labelTextProps: {
            as: 'div',
            className: 'flex flex-col gap-1',
          },
        }}
        className="items-start"
      />

      {value.mode !== 'none' ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Default mode
          </Text>
          <ToggleButton<TAppBranchRunPreviewMode>
            value={value.mode}
            onChange={(mode) => onChange({ ...value, mode })}
            options={[
              {
                value: 'build-only',
                label: previewModeDisplayLabel('build-only'),
              },
              {
                value: 'plan-only',
                label: previewModeDisplayLabel('plan-only'),
              },
              { value: 'apply', label: previewModeDisplayLabel('apply') },
            ]}
          />
        </div>
      ) : null}

      {value.mode !== 'none' && value.mode !== 'build-only' ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            Default install
          </Text>
          <Select
            options={installOptions}
            value={value.installId}
            onChange={(installId) =>
              onChange({ ...value, installId, installTargetMode: 'install' })
            }
            placeholder="Select an install"
            disabled={disabled || installOptions.length === 0}
            menuPlacement="bottom"
          />
        </div>
      ) : null}

      {hasGithubVCS && value.mode !== 'none' ? (
        <div className="flex flex-col gap-2">
          <CheckboxInput
            id="preview-set-statuses"
            checked={value.setStatuses}
            onChange={(e) =>
              onChange({ ...value, setStatuses: e.target.checked })
            }
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
          <CheckboxInput
            id="preview-ignore-drafts"
            checked={value.ignoreDrafts}
            onChange={(e) =>
              onChange({ ...value, ignoreDrafts: e.target.checked })
            }
            disabled={disabled}
            labelProps={{ labelText: 'Ignore draft pull requests' }}
          />
        </div>
      ) : null}
    </div>
  )
}
