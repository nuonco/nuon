'use client'

import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import type { TBuild } from '@/types'

export const InstallComponentBuildCard = ({ build }: { build: TBuild }) => {
  const { install } = useInstall()

  return (
    <ContextTooltip
      width="w-fit"
      title="Component build"
      items={[
        {
          id: `build-id-`,
          title: 'Build ID',
          subtitle: <ID variant="label">{build?.id}</ID>,
        },
        {
          id: `build-date-`,
          title: 'Build date',
          subtitle: (
            <Time time={build?.created_at} variant="label" theme="neutral" />
          ),
        },
        {
          id: `build-duration-`,
          title: 'Build duration',
          subtitle: (
            <Duration
              beginTime={build?.created_at}
              endTime={build?.updated_at}
              variant="label"
              theme="neutral"
            />
          ),
        },
      ]}
    >
      <Card className="!p-2 !flex-row">
        <Text weight="strong">
          <Link
            href={`/${install?.org_id}/apps/${install?.app_id}/configs/${install?.app_config_id}/components/${build?.component_id}/builds/${build.id}`}
          >
            Component build <Icon variant="Question" />
          </Link>
        </Text>
        <Status status={build?.status_v2?.status} variant="badge" />
      </Card>
    </ContextTooltip>
  )
}

export const InstallComponentBuildCardSkeleton = () => {
  return <Skeleton height="42px" width="240px" />
}
