import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds } from '@/lib'

export const InstallImageSummary = ({
  componentId,
  sourceRef,
}: {
  componentId: string
  sourceRef?: string
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: builds, isLoading } = useQuery({
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

  const latestBuild = builds?.data?.[0]
  const build = builds?.data?.find((b) => !!b.source_digest) ?? latestBuild

  if (!isLoading && !build) {
    return (
      <EmptyState
        variant="table"
        size="sm"
        emptyTitle="No builds yet"
        emptyMessage="Build this component to publish an image."
      />
    )
  }

  const buildHref =
    build?.id && install?.app_id
      ? `/${org?.id}/apps/${install.app_id}/components/${componentId}/builds/${build.id}`
      : undefined

  return (
    <Card className="!p-4 !gap-4">
      <div className="grid gap-4 md:grid-cols-2">
        <LabeledValue label="Status" loading={isLoading}>
          <Status status={build?.status_v2?.status ?? build?.status} />
        </LabeledValue>
        <LabeledValue label="Built" loading={isLoading}>
          <Time
            variant="subtext"
            time={build?.resolved_at ?? build?.created_at}
            format="relative"
          />
        </LabeledValue>
        <LabeledValue label="Source ref" loading={isLoading}>
          <Text variant="subtext" family="mono" className="break-all">
            {build?.source_ref ?? build?.source_image ?? sourceRef ?? '—'}
          </Text>
        </LabeledValue>
        <LabeledValue label="Resolved tag" loading={isLoading}>
          <Text variant="subtext" family="mono">
            {build?.resolved_tag ?? '—'}
          </Text>
        </LabeledValue>
        <LabeledValue label="Digest" className="md:col-span-2">
          {build?.source_digest ? (
            <ClickToCopy>
              <Text variant="subtext" family="mono" className="break-all">
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
      {buildHref ? <Link href={buildHref}>View build</Link> : null}
    </Card>
  )
}
