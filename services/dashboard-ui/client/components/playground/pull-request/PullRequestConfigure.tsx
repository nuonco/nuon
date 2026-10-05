import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Toggle } from '@/components/common/form/Toggle'
import { PageContent } from '@/components/layout/PageContent'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import type { TAppBranchPreviewConfig, TAppBranchRunPreviewMode } from '@/types'
import { PullRequestHeader } from './PullRequestHeader'
import {
  installCandidates,
  type TPlaygroundInstallCandidate,
  type TPlaygroundPullRequest,
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

export interface IPullRequestConfigure {
  pullRequest: TPlaygroundPullRequest
  installs?: TPlaygroundInstallCandidate[]
  onSave: (override: TAppBranchPreviewConfig) => void
  onCancel?: () => void
}

export const PullRequestConfigure = ({
  pullRequest,
  installs = installCandidates,
  onSave,
  onCancel,
}: IPullRequestConfigure) => {
  const resolved = pullRequest.resolved_preview_config
  const [mode, setMode] = useState<TAppBranchRunPreviewMode>(
    resolved.mode ?? 'plan-only'
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

  const save = () =>
    onSave({
      mode,
      install_id: needsInstall ? installId : undefined,
      install_name: needsInstall ? selectedInstall?.name : undefined,
      set_statuses: setStatuses,
      comment,
      ignore_drafts: ignoreDrafts,
    })

  return (
    <PageContent>
      <PageSection>
        <PullRequestHeader pullRequest={pullRequest} />
      </PageSection>

      <PageSection className="pt-0">
        <Card>
          <SectionHeader
            title="Preview mode"
            description="Applies to every preview run on this pull request, overriding the branch default."
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
                        {option.description}
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
            description="What Nuon writes back to this pull request."
          />
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
        </Card>

        <div className="flex flex-wrap items-center gap-2">
          <Button variant="primary" size="lg" onClick={save} disabled={!isValid}>
            Save &amp; re-run
          </Button>
          {onCancel ? (
            <Button variant="ghost" size="lg" onClick={onCancel}>
              Cancel
            </Button>
          ) : null}
          {isValid ? (
            <span className="flex items-center gap-1">
              <Icon variant="InfoIcon" size={12} theme="neutral" />
              <Text variant="subtext" theme="neutral">
                Runs on every push to
              </Text>
              <Text variant="subtext" family="mono" theme="neutral">
                {pullRequest.head_ref}
              </Text>
            </span>
          ) : (
            <Text variant="subtext" theme="warn">
              Select an install to use {previewModeDisplayLabel(mode)}.
            </Text>
          )}
        </div>
      </PageSection>
    </PageContent>
  )
}
