import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Toggle } from '@/components/common/form/Toggle'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import type { TAppBranchPreviewConfig, TAppBranchRunPreviewMode } from '@/types'
import type {
  TPlaygroundInstallCandidate,
  TPlaygroundPullRequest,
} from './fixtures'

const MODES: {
  value: TAppBranchRunPreviewMode
  description: string
  needsInstall: boolean
}[] = [
  {
    value: 'build-only',
    description:
      'Validate the config and build every component. Nothing is planned or applied, so no install is needed.',
    needsInstall: false,
  },
  {
    value: 'plan-only',
    description:
      'Build, then plan against an install to see what would change. Nothing is applied.',
    needsInstall: true,
  },
  {
    value: 'apply',
    description:
      'Build, plan, and apply to an install on every push to this PR.',
    needsInstall: true,
  },
]

const InstallOption = ({
  install,
  checked,
  onSelect,
}: {
  install: TPlaygroundInstallCandidate
  checked: boolean
  onSelect: () => void
}) => (
  <RadioInput
    name="preview-install"
    value={install.id}
    checked={checked}
    onChange={onSelect}
    labelProps={{
      className: checked
        ? 'border rounded-md !border-primary-500 bg-primary-50 dark:bg-primary-950'
        : 'border rounded-md',
      labelText: (
        <span className="flex flex-wrap items-center justify-between gap-2 w-full">
          <span className="flex items-center gap-2">
            <Text variant="body" family="mono">
              {install.name}
            </Text>
            <Badge size="xs" theme="neutral">
              {install.cloud_platform} · {install.region}
            </Badge>
          </span>
          <span className="flex flex-wrap gap-1">
            {Object.entries(install.labels).map(([key, value]) => (
              <LabelBadge
                key={key}
                labelKey={key}
                labelValue={value}
                size="sm"
              />
            ))}
          </span>
        </span>
      ),
    }}
  />
)

export interface IPullRequestConfigure extends IModal {
  pullRequest: TPlaygroundPullRequest
  installs?: TPlaygroundInstallCandidate[]
  onSave: (override: TAppBranchPreviewConfig) => void | Promise<unknown>
  modeDescription?: string
  githubLocked?: boolean
  isSaving?: boolean
}

export const PullRequestConfigure = ({
  pullRequest,
  installs = [],
  onSave,
  modeDescription = 'Applies to every preview run on this pull request, overriding the branch default.',
  githubLocked = false,
  isSaving = false,
  modalId,
  ...modalProps
}: IPullRequestConfigure) => {
  const { removeModal } = useSurfaces()
  const [saving, setSaving] = useState(false)
  const busy = isSaving || saving
  const resolved = pullRequest.resolved_preview_config
  const [mode, setMode] = useState<TAppBranchRunPreviewMode>(
    resolved.mode && resolved.mode !== 'none' ? resolved.mode : 'plan-only'
  )
  const [installId, setInstallId] = useState<string | undefined>(
    resolved.install_id
  )
  const [setStatuses, setSetStatuses] = useState(resolved.set_statuses ?? true)
  const [comment, setComment] = useState(resolved.comment ?? true)
  const [ignoreDrafts, setIgnoreDrafts] = useState(
    resolved.ignore_drafts ?? true
  )

  const needsInstall = MODES.find((m) => m.value === mode)?.needsInstall ?? false
  const selectedInstall = installs.find((install) => install.id === installId)
  const isValid = !needsInstall || Boolean(installId)

  const save = async () => {
    setSaving(true)
    try {
      await onSave({
        mode,
        install_id: needsInstall ? installId : undefined,
        install_name: needsInstall ? selectedInstall?.name : undefined,
        set_statuses: setStatuses,
        comment,
        ignore_drafts: ignoreDrafts,
      })
      removeModal(modalId)
    } catch {
    } finally {
      setSaving(false)
    }
  }

  const hint =
    isValid && !githubLocked ? (
      <span className="flex items-center gap-1">
        <Icon variant="InfoIcon" size={12} theme="neutral" />
        <Text variant="subtext" theme="neutral">
          Runs on every push to
        </Text>
        <Text variant="subtext" family="mono" theme="neutral">
          {pullRequest.head_ref}
        </Text>
      </span>
    ) : isValid ? null : (
      <Text variant="subtext" theme="warn">
        Select an install to use {previewModeDisplayLabel(mode)}.
      </Text>
    )

  const fields = (
    <>
        <Card>
          <SectionHeader
            title="Preview mode"
            description={modeDescription}
          />
          <div className="flex flex-col gap-1">
            {MODES.map((option) => (
              <RadioInput
                key={option.value}
                name="preview-mode"
                value={option.value}
                checked={mode === option.value}
                onChange={() => setMode(option.value)}
                labelProps={{
                  className: 'items-start',
                  labelText: (
                    <span className="flex flex-col gap-0.5">
                      <Text variant="body" weight="strong">
                        {previewModeDisplayLabel(option.value)}
                      </Text>
                      <Text variant="subtext" theme="neutral">
                        {githubLocked && option.value === 'apply'
                          ? 'Build, plan, and apply this run to an install.'
                          : option.description}
                      </Text>
                    </span>
                  ),
                }}
              />
            ))}
          </div>
        </Card>

        {needsInstall ? (
          <Card>
            <SectionHeader
              title="Preview install"
              description="Which install this PR previews against. Required for plan and apply."
              status={
                installId ? null : (
                  <Badge size="sm" theme="warn">
                    required
                  </Badge>
                )
              }
            />
            <div className="flex flex-col gap-1">
              {installs.map((install) => (
                <InstallOption
                  key={install.id}
                  install={install}
                  checked={installId === install.id}
                  onSelect={() => setInstallId(install.id)}
                />
              ))}
            </div>
          </Card>
        ) : null}

        <Card>
          <SectionHeader
            title="GitHub"
            description={
              githubLocked
                ? 'These follow the branch preview config. A re-run does not change them.'
                : 'What Nuon writes back to this pull request.'
            }
          />
          {githubLocked ? (
            <span className="flex flex-wrap gap-1">
              {comment ? (
                <Badge size="xs" theme="neutral">
                  comment
                </Badge>
              ) : null}
              {setStatuses ? (
                <Badge size="xs" theme="neutral">
                  statuses
                </Badge>
              ) : null}
              {ignoreDrafts ? (
                <Badge size="xs" theme="neutral">
                  ignore drafts
                </Badge>
              ) : null}
            </span>
          ) : (
            <div className="flex flex-col gap-3">
              <Toggle
                checked={comment}
                onChange={setComment}
                label="Update the preview comment"
                description="Edit a single sticky comment as each run progresses."
              />
              <Toggle
                checked={setStatuses}
                onChange={setSetStatuses}
                label="Set commit statuses"
                description="Report each preview run as a commit status on the head SHA."
              />
              <Toggle
                checked={ignoreDrafts}
                onChange={setIgnoreDrafts}
                label="Ignore draft pull requests"
                description="Skip preview runs while this PR is a draft."
              />
            </div>
          )}
        </Card>

        {hint}
    </>
  )

  return (
    <Modal
      heading="Edit preview config"
      size="lg"
      modalId={modalId}
      primaryActionTrigger={{
        children: busy ? 'Saving...' : 'Save & re-run',
        disabled: !isValid || busy,
        onClick: () => void save(),
        variant: 'primary',
      }}
      {...modalProps}
    >
      {fields}
    </Modal>
  )
}
