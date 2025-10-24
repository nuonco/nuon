import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { getHelmOutputStatus } from '@/utils/helm-utils'

export const Overview = ({
  createdAt,
  outputs,
}: {
  createdAt: string
  outputs: Record<string, any>
}) => {
  const firstDeployment = Object.values(
    Object.values(outputs.deployments)[0] || {}
  )[0] as any
  const releaseName =
    firstDeployment?.metadata?.annotations?.['meta.helm.sh/release-name'] ||
    'Unknown'
  const namespace =
    firstDeployment?.metadata?.annotations?.[
      'meta.helm.sh/release-namespace'
    ] || 'Unknown'

  return (
    <div className="flex flex-col gap-6">
      <HeadingGroup>
        <Text
          className="flex items-center gap-4"
          variant="base"
          weight="strong"
        >
          Helm outputs{' '}
          <Status status={getHelmOutputStatus(outputs)} variant="badge" />
        </Text>
        <Text className="flex items-center gap-6" theme="neutral">
          <Text variant="subtext">
            Release: <b>{releaseName}</b>
          </Text>{' '}
          <Text variant="subtext">
            Namespace: <b>{namespace}</b>
          </Text>{' '}
          <Text variant="subtext">
            Created:{' '}
            <Time time={createdAt} variant="subtext" weight="stronger" />
          </Text>
        </Text>
      </HeadingGroup>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
        <OutputCard
          count={Object.keys(outputs?.deployments)?.length}
          title="Deployments"
        />
        <OutputCard
          count={Object.keys(outputs?.services)?.length}
          title="Services"
        />
        <OutputCard
          count={Object.keys(outputs?.ingresses)?.length}
          title="Ingresses"
        />
        <OutputCard
          count={Object.keys(outputs?.resources)?.length}
          title="Resources"
        />
      </div>
    </div>
  )
}

const OutputCard = ({
  count,

  title,
}: {
  count: number
  title: string
}) => {
  return (
    <Card className="!gap-2">
      <Text weight="strong" theme="neutral">
        {title}
      </Text>
      <Text variant="h3" weight="stronger">
        {count}
      </Text>
    </Card>
  )
}
