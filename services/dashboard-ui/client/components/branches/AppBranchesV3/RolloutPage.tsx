import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import {
  CommitLink,
  PullRequestLink,
  TagLink,
} from '@/components/common/GitReferenceLink'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Panel } from '@/components/surfaces/Panel'
import { ConfigChangesViewer } from './ConfigChanges'
import type { TLatestRollout } from './fixtures'
import { InstallGroupCards } from './InstallGroupCards'
import { RolloutTimeline } from './RolloutTimeline'

const TRIGGER_LABEL: Record<TLatestRollout['source']['kind'], string> = {
  'pull-request': 'Pull request',
  tag: 'Tag',
  manual: 'Manual run',
  commit: 'Push',
}

const SourceLine = ({ source }: { source: TLatestRollout['source'] }) => {
  if (source.kind === 'pull-request') {
    return (
      <span className="flex flex-wrap items-center gap-2">
        <PullRequestLink
          number={source.number}
          href={source.url}
          label={`Pull request #${source.number}`}
          weight="strong"
        />
        {source.baseBranch ? (
          <Text variant="subtext" theme="neutral">
            into{' '}
            <Text as="span" variant="subtext" family="mono">
              {source.baseBranch}
            </Text>
          </Text>
        ) : null}
      </span>
    )
  }
  if (source.kind === 'tag') {
    return <TagLink tag={source.tag} href={source.url} weight="strong" />
  }
  return null
}

const CommitInfo = ({ rollout }: { rollout: TLatestRollout }) => {
  const [subject, ...body] = rollout.commit.message.split('\n')
  const description = body.join('\n').trim()

  return (
    <div className="flex min-w-0 flex-col gap-2">
      <SourceLine source={rollout.source} />
      <Text variant="body" weight="strong">
        {subject}
      </Text>
      {description ? (
        <Text
          as="p"
          variant="subtext"
          theme="neutral"
          className="whitespace-pre-line break-words"
        >
          {description}
        </Text>
      ) : null}
      <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
        <CommitLink sha={rollout.commit.sha} href={rollout.commit.shaUrl} />
        <Text variant="subtext" theme="neutral">
          {rollout.commit.author}
        </Text>
        <Time
          variant="subtext"
          theme="neutral"
          time={rollout.commit.createdAt}
          format="relative"
        />
        <Text variant="subtext" theme="neutral">
          · {TRIGGER_LABEL[rollout.source.kind]}
        </Text>
      </span>
    </div>
  )
}

export interface IRolloutPage {
  rollout: TLatestRollout
  backHref?: string
}

export const RolloutPage = ({ rollout, backHref }: IRolloutPage) => (
  <PageSection>
    <div className="flex flex-col gap-3">
      {backHref ? (
        <Link href={backHref}>
          <Icon variant="ArrowLeftIcon" />
          Previous runs
        </Link>
      ) : null}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="flex flex-wrap items-center gap-3">
          <Text variant="h3" weight="stronger" level={2}>
            Rollout
          </Text>
          <Status status={rollout.status} />
          <Text variant="subtext" theme="neutral" family="mono">
            {rollout.configChanges.versionLabel}
          </Text>
        </span>
        <Panel
          panelKey="rollout-config-changes"
          size="3/4"
          heading="Template changes"
          triggerButton={{
            variant: 'secondary',
            className: 'gap-3',
            children: (
              <>
                Template changes
                <ChangeCountSummary
                  added={rollout.configChanges.summary.added}
                  updated={rollout.configChanges.summary.changed}
                  removed={rollout.configChanges.summary.removed}
                />
              </>
            ),
          }}
        >
          <ConfigChangesViewer
            sections={rollout.configChanges.sections}
            files={rollout.configChanges.files}
            versionLabel={rollout.configChanges.versionLabel}
            previousSha={rollout.commit.previousSha}
            sha={rollout.commit.sha}
          />
        </Panel>
      </div>
      <RolloutTimeline status={rollout.status} />
    </div>
    <CommitInfo rollout={rollout} />
    <InstallGroupCards groups={rollout.installGroups} commit={rollout.commit} />
  </PageSection>
)
