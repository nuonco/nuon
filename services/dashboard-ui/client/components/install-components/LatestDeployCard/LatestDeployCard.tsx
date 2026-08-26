import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TDeploy } from '@/types'

export interface ILatestDeployCard {
  deploy?: TDeploy
  href?: string
  isLoading?: boolean
}

export const LatestDeployCard = ({
  deploy,
  href,
  isLoading,
}: ILatestDeployCard) => {
  if (!isLoading && !deploy) {
    return (
      <EmptyState
        variant="table"
        size="sm"
        emptyTitle="No deploys yet"
        emptyMessage="Deploys appear here once this component is deployed to the install."
      />
    )
  }

  return (
    <Card>
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
            {deploy?.install_deploy_type === 'teardown' ? 'Teardown' : 'Deploy'}
          </Text>
        </LabeledValue>
      </div>
      {href ? <Link href={href}>View deploy</Link> : null}
    </Card>
  )
}
