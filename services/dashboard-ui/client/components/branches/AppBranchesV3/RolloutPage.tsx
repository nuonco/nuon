import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { BranchOverview } from '@/components/branches/BranchOverview/BranchOverview'
import {
  groupStageStatus,
  type TOverviewStage,
  type TOverviewStageStatus,
} from '@/components/branches/BranchOverview/overview-loading'
import type { TTrackGroup } from '@/components/branches/BranchOverview/RolloutTrack'
import { TemplateChangesButton } from '@/components/branches/ConfigChanges'
import type { TLatestRollout, TRolloutInstallGroup } from './fixtures'

const toTrackGroup = (group: TRolloutInstallGroup): TTrackGroup => ({
  id: group.id ?? group.name ?? 'group',
  name: group.name || 'Install group',
  status: group.status,
  maxParallel: group.max_parallel ?? 1,
  approval: group.auto_approve_on_policies_passing
    ? 'Auto-approves when policies pass'
    : 'Manual approval',
  match: group.default
    ? { kind: 'default' }
    : group.label_selector?.match_labels
      ? { kind: 'labels', labels: group.label_selector.match_labels }
      : { kind: 'pinned' },
  installs: group.installs.map((install) => ({
    id: install.id,
    name: install.name,
    status: install.awaitingApproval ? 'approval-awaiting' : install.status,
  })),
})

const stagesFor = (status: string): TOverviewStage[] => {
  const rollout = groupStageStatus(status)
  const done: TOverviewStageStatus =
    rollout === 'pending' ? 'pending' : 'success'
  return [
    { id: 'starting', label: 'Starting workflow', status: 'success' as const },
    { id: 'fetch-commit', label: 'Fetch commit', status: done },
    {
      id: 'app-config',
      label: 'Compile and diff app config',
      status: done,
    },
    { id: 'build-components', label: 'Build components', status: done },
    { id: 'rollout', label: 'Rollout', status: rollout },
  ]
}

export interface IRolloutPage {
  rollout: TLatestRollout
  backHref?: string
}

export const RolloutPage = ({ rollout, backHref }: IRolloutPage) => (
  <PageSection>
    {backHref ? (
      <Link href={backHref}>
        <Icon variant="ArrowLeftIcon" />
        Previous runs
      </Link>
    ) : null}
    <BranchOverview
      hasPlan
      groups={rollout.installGroups.map(toTrackGroup)}
      rolloutHref="#rollout"
      groupHref={(id) => `#groups/${id}`}
      loadingStages={stagesFor(rollout.status)}
      versionLabel={rollout.configChanges.versionLabel}
      previousSha={rollout.commit.previousSha}
      rollout={{
        id: rollout.id,
        href: '#run',
        source: rollout.source,
        title: rollout.title,
        sha: rollout.commit.sha,
        shaUrl: rollout.commit.shaUrl,
        author: rollout.commit.author,
        status: rollout.status,
        commit: {
          message: rollout.commit.message,
          author: rollout.commit.author,
          sha: rollout.commit.sha,
          shaUrl: rollout.commit.shaUrl,
          createdAt: rollout.commit.createdAt,
        },
      }}
      changes={
        <TemplateChangesButton
          summary={rollout.configChanges.summary}
          sections={rollout.configChanges.sections}
          files={rollout.configChanges.files}
          versionLabel={rollout.configChanges.versionLabel}
          previousSha={rollout.commit.previousSha}
          sha={rollout.commit.sha}
        />
      }
    />
  </PageSection>
)
