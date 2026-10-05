import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { PullRequestHeader } from './PullRequestHeader'
import { PullRequestRunsTable } from './PullRequestRunsTable'
import {
  previewRuns,
  type TPlaygroundPullRequest,
  type TPlaygroundPullRequestRun,
} from './fixtures'

const CONFIG_SOURCE_LABEL: Record<
  TPlaygroundPullRequest['config_status'],
  { label: string; theme: 'brand' | 'neutral' | 'warn'; hint: string }
> = {
  configured: {
    label: 'PR override',
    theme: 'brand',
    hint: 'Set on this pull request. Branch defaults are ignored.',
  },
  'ready-from-defaults': {
    label: 'Branch default',
    theme: 'neutral',
    hint: 'Inherited from the branch config. Override it to change just this PR.',
  },
  'needs-input': {
    label: 'Needs an install',
    theme: 'warn',
    hint: 'The branch default has no install target, so runs fall back to build-only.',
  },
}

export interface IPullRequestDetail {
  pullRequest: TPlaygroundPullRequest
  runs?: TPlaygroundPullRequestRun[]
  onEditConfig: () => void
  onRerun?: () => void
  isLoading?: boolean
}

export const PullRequestDetail = ({
  pullRequest,
  runs = previewRuns,
  onEditConfig,
  onRerun,
  isLoading = false,
}: IPullRequestDetail) => {
  const resolved = pullRequest.resolved_preview_config
  const source = CONFIG_SOURCE_LABEL[pullRequest.config_status]
  const isClosed = pullRequest.state !== 'open'
  const needsInput = pullRequest.config_status === 'needs-input'

  return (
    <PageContent>
      <PageSection>
        <PullRequestHeader
          pullRequest={pullRequest}
          actions={
            isClosed ? null : (
              <>
                {onRerun ? (
                  <Button variant="secondary" size="md" onClick={onRerun}>
                    Re-run preview
                  </Button>
                ) : null}
                <Button
                  variant={needsInput ? 'primary' : 'secondary'}
                  size="md"
                  onClick={onEditConfig}
                >
                  {needsInput ? 'Choose an install' : 'Edit preview config'}
                </Button>
              </>
            )
          }
        />
      </PageSection>

      <PageSection className="pt-0">
        <Card>
          <SectionHeader
            title="Preview configuration"
            description={source.hint}
            status={
              <Badge size="sm" theme={source.theme}>
                {source.label}
              </Badge>
            }
            actions={
              isClosed ? null : (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onEditConfig}
                  tooltipProps={{
                    position: 'top',
                    tipContent: 'Edit this PR’s preview config',
                  }}
                >
                  <Icon variant="PencilSimpleIcon" size={12} />
                </Button>
              )
            }
          />
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
            <LabeledValue label="Mode">
              <Badge size="sm">
                {previewModeDisplayLabel(resolved.mode ?? 'plan-only')}
              </Badge>
            </LabeledValue>
            <LabeledValue label="Install">
              {resolved.mode === 'build-only' ? (
                <Text variant="subtext" theme="neutral">
                  Not used
                </Text>
              ) : resolved.install_name ? (
                <Text variant="subtext" family="mono">
                  {resolved.install_name}
                </Text>
              ) : (
                <Text variant="subtext" theme="warn">
                  Not set
                </Text>
              )}
            </LabeledValue>
            <LabeledValue label="GitHub">
              <span className="flex flex-wrap gap-1">
                {resolved.comment ? (
                  <Badge size="xs" theme="neutral">
                    comment
                  </Badge>
                ) : null}
                {resolved.set_statuses ? (
                  <Badge size="xs" theme="neutral">
                    statuses
                  </Badge>
                ) : null}
                {resolved.ignore_drafts ? (
                  <Badge size="xs" theme="neutral">
                    ignore drafts
                  </Badge>
                ) : null}
              </span>
            </LabeledValue>
          </div>
        </Card>

        <SectionHeader
          title="Preview runs"
          description={`Every preview run on pull request #${pullRequest.number}, newest first.`}
          status={
            <Badge size="sm" theme="neutral">
              {runs.length}
            </Badge>
          }
        />
        <PullRequestRunsTable
          pullRequest={pullRequest}
          runs={runs}
          isLoading={isLoading}
        />
      </PageSection>
    </PageContent>
  )
}
