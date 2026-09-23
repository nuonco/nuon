import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TComponentBuild, TDeploy } from '@/types'

export type TLatestDeployCardVariant = 'deploy' | 'sync'

const COPY = {
  deploy: {
    emptyTitle: 'No deploys yet',
    emptyMessage:
      'Deploys appear here once this component is deployed to the install.',
    applyType: 'Deploy',
    view: 'View deploy',
  },
  sync: {
    emptyTitle: 'No syncs yet',
    emptyMessage:
      'Syncs appear here once this image is synced to the install.',
    applyType: 'Sync',
    view: 'View sync',
  },
} as const

export interface ILatestDeployCard {
  build?: TComponentBuild
  buildHref?: string
  deploy?: TDeploy
  flush?: boolean
  href?: string
  isLoading?: boolean
  variant?: TLatestDeployCardVariant
}

export const LatestDeployCard = ({
  build,
  buildHref,
  deploy,
  flush = false,
  href,
  isLoading,
  variant = 'deploy',
}: ILatestDeployCard) => {
  const commit = build?.vcs_connection_commit
  const copy = COPY[variant]
  if (!isLoading && !deploy) {
    return (
      <EmptyState
        variant="table"
        size="sm"
        emptyTitle={copy.emptyTitle}
        emptyMessage={copy.emptyMessage}
      />
    )
  }

  const Wrapper = flush ? 'div' : Card

  return (
    <Wrapper className={flush ? 'flex flex-col gap-4' : undefined}>
      <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
        <LabeledStatus
          label="Status"
          loading={isLoading}
          statusProps={{ status: deploy?.status_v2?.status }}
          tooltipProps={{
            tipContent: deploy?.status_v2?.status_human_description,
            position: 'bottom',
          }}
        />
        <LabeledValue label="Started" loading={isLoading}>
          <Time variant="subtext" time={deploy?.created_at} format="relative" />
        </LabeledValue>
        <LabeledValue label="Duration" loading={isLoading}>
          <Duration
            variant="subtext"
            beginTime={deploy?.created_at}
            endTime={deploy?.updated_at}
          />
        </LabeledValue>
        <LabeledValue label="Type" loading={isLoading}>
          <Text variant="subtext">
            {deploy?.install_deploy_type === 'teardown'
              ? 'Teardown'
              : copy.applyType}
          </Text>
        </LabeledValue>
      </div>
      {build ? (
        <div className="flex flex-col gap-4 border-t pt-4">
          <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
            <LabeledStatus
              label="Build"
              statusProps={{ status: build.status_v2?.status ?? build.status }}
              tooltipProps={{
                tipContent: build.status_v2?.status_human_description,
                position: 'bottom',
              }}
            />
            <LabeledValue label="Built">
              <Time
                variant="subtext"
                time={build.created_at}
                format="relative"
              />
            </LabeledValue>
            <LabeledValue label="Build ID">
              {buildHref ? (
                <Link href={buildHref} className="font-mono">
                  {build.id}
                </Link>
              ) : (
                <ID>{build.id}</ID>
              )}
            </LabeledValue>
            {build.no_op ? (
              <LabeledValue label="Artifact">
                <Text variant="subtext">Reused</Text>
              </LabeledValue>
            ) : null}
          </div>
          {commit ? (
            <BranchRunCommit
              showStatus={false}
              sha={commit.sha}
              message={commit.message?.split('\n')[0]}
              author={commit.author_name}
              avatarUrl={commit.author_avatar_url}
            />
          ) : null}
        </div>
      ) : null}
      {href ? <Link href={href}>{copy.view}</Link> : null}
    </Wrapper>
  )
}
