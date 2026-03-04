import { Card } from '@/components/common/Card'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export const DeploymentStatuses = ({
  deployments,
}: {
  deployments: Record<string, any>
}) => {
  const allDeployments = Object.values(deployments).flatMap((namespace) =>
    Object.entries(namespace as any)
  )

  return (
    <Card>
      <Text weight="strong">Deployment statuses</Text>

      <div className="flex flex-col gap-2">
        {allDeployments.map(([name, deployment]: [string, any]) => {
          const status = deployment.status
          const replicas = {
            desired: status?.replicas || 0,
            ready: status?.readyReplicas || 0,
            available: status?.availableReplicas || 0,
          }

          const isHealthy =
            replicas.ready === replicas.desired &&
            replicas.available === replicas.desired

          const healthPercentage =
            replicas.desired > 0
              ? Math.round((replicas.ready / replicas.desired) * 100)
              : 0

          return (
            <div key={name} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Status
                  status={isHealthy ? 'healthy' : 'info'}
                  isWithoutText
                  variant="timeline"
                />
                <Text variant="body" weight="strong">
                  {name}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {replicas.ready}/{replicas.desired} replicas
                </Text>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-24 bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full transition-all duration-500 ${
                      isHealthy
                        ? 'bg-green-600 dark:bg-green-400'
                        : 'bg-blue-600 dark:bg-blue-400'
                    }`}
                    style={{ width: `${healthPercentage}%` }}
                  />
                </div>
                <Text variant="subtext" weight="strong" theme="neutral">
                  {healthPercentage}%
                </Text>
              </div>
            </div>
          )
        })}
      </div>
    </Card>
  )
}
