import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Badge } from '@/components/common/Badge'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds, getInstallComponent } from '@/lib'
import type { TComponentBuild, TDeploy } from '@/types'

export interface IInstallImageSummary {
  build?: TComponentBuild
  buildHref?: string
  buildLoading?: boolean
  deploy?: TDeploy
  sourceRef?: string
  syncHref?: string
  syncLoading?: boolean
}

export const InstallImageSummary = ({
  build,
  buildHref,
  buildLoading = false,
  deploy,
  sourceRef,
  syncHref,
  syncLoading = false,
}: IInstallImageSummary) => {
  const commit =
    build?.vcs_connection_commit ?? build?.app_branch_run?.vcs_connection_commit
  const showSyncEmpty = !syncLoading && !deploy
  const showBuildEmpty = !buildLoading && !build

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Text variant="subtext" weight="strong" theme="neutral">
          Latest sync
        </Text>
        {showSyncEmpty ? (
          <EmptyState
            variant="table"
            size="sm"
            emptyTitle="No syncs yet"
            emptyMessage="Syncs appear here once this image is synced to the install."
          />
        ) : (
          <div className="flex flex-col gap-4">
            <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
              <LabeledStatus
                label="Status"
                loading={syncLoading}
                statusProps={{ status: deploy?.status_v2?.status }}
                tooltipProps={{
                  tipContent: deploy?.status_v2?.status_human_description,
                  position: 'bottom',
                }}
              />
              <LabeledValue label="Started" loading={syncLoading}>
                <Time
                  variant="subtext"
                  time={deploy?.created_at}
                  format="relative"
                />
              </LabeledValue>
              <LabeledValue label="Duration" loading={syncLoading}>
                <Duration
                  variant="subtext"
                  beginTime={deploy?.created_at}
                  endTime={deploy?.updated_at}
                />
              </LabeledValue>
              <LabeledValue label="Type" loading={syncLoading}>
                <Text variant="subtext">
                  {deploy?.install_deploy_type === 'teardown'
                    ? 'Teardown'
                    : 'Sync'}
                </Text>
              </LabeledValue>
            </div>
            {syncHref ? <Link href={syncHref}>View sync</Link> : null}
          </div>
        )}
      </div>

      <div className="flex flex-col gap-2 border-t pt-4">
        <Text variant="subtext" weight="strong" theme="neutral">
          Current build
        </Text>
        {showBuildEmpty ? (
          <EmptyState
            variant="table"
            size="sm"
            emptyTitle="No builds yet"
            emptyMessage="Build this component to publish an image."
          />
        ) : (
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 md:grid-cols-2">
              <LabeledStatus
                label="Status"
                loading={buildLoading}
                statusProps={{
                  status: build?.status_v2?.status ?? build?.status,
                }}
                tooltipProps={{
                  tipContent: build?.status_v2?.status_human_description,
                  position: 'bottom',
                }}
              />
              <LabeledValue label="Built" loading={buildLoading}>
                <Time
                  variant="subtext"
                  time={build?.resolved_at ?? build?.created_at}
                  format="relative"
                />
              </LabeledValue>
              <LabeledValue label="Source ref" loading={buildLoading}>
                <Text variant="subtext" family="mono" className="break-all">
                  {build?.source_ref ?? build?.source_image ?? sourceRef ?? '—'}
                </Text>
              </LabeledValue>
              <LabeledValue label="Resolved tag" loading={buildLoading}>
                <Text variant="subtext" family="mono">
                  {build?.resolved_tag ?? '—'}
                </Text>
              </LabeledValue>
              <LabeledValue
                label="Digest"
                loading={buildLoading}
                className="md:col-span-2"
              >
                {build?.source_digest ? (
                  <ClickToCopy>
                    <Text
                      variant="subtext"
                      family="mono"
                      className="break-all"
                    >
                      {build.source_digest}
                    </Text>
                  </ClickToCopy>
                ) : (
                  <Text variant="subtext" theme="neutral">
                    Not recorded for this build
                  </Text>
                )}
              </LabeledValue>
              {build?.no_op ? (
                <LabeledValue label="Rebuild">
                  <Badge size="sm" variant="code" theme="neutral">
                    no-op
                  </Badge>
                </LabeledValue>
              ) : null}
            </div>
            {commit ? (
              <BranchRunCommit
                showStatus={false}
                href={buildHref}
                sha={commit.sha}
                message={commit.message?.split('\n')[0]}
                author={commit.author_name}
                avatarUrl={commit.author_avatar_url}
              />
            ) : null}
            {buildHref ? <Link href={buildHref}>View build</Link> : null}
          </div>
        )}
      </div>
    </div>
  )
}

export const InstallImageSummaryContainer = ({
  componentId,
  sourceRef,
}: {
  componentId: string
  sourceRef?: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent, isLoading: syncLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const { data: builds, isLoading: buildLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component-builds', org?.id, componentId, 0],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: 10,
        offset: 0,
      }),
    refetchInterval: 20000,
    enabled: !!org?.id && !!componentId,
  })

  const latestDeploy = installComponent?.install_deploys?.[0]
  const latestBuild = builds?.data?.[0]
  const build = builds?.data?.find((item) => !!item.source_digest) ?? latestBuild

  return (
    <InstallImageSummary
      build={build}
      buildHref={
        build?.id && install?.app_id
          ? `/${org?.id}/apps/${install.app_id}/components/${componentId}/builds/${build.id}`
          : undefined
      }
      buildLoading={buildLoading}
      deploy={latestDeploy}
      sourceRef={sourceRef}
      syncHref={
        latestDeploy?.id
          ? `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${latestDeploy.id}`
          : undefined
      }
      syncLoading={syncLoading}
    />
  )
}
