import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export const DeploymentsDetails = ({
  deployments,
}: {
  deployments: Record<string, any>
}) => {
  const hasDeployments = Object.keys(deployments).length > 0

  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Deployments details</Text>

      {hasDeployments ? (
        Object.entries(deployments).map(([namespace, namespaceDeployments]) => (
          <Card key={namespace} className="!p-0 !gap-0">
            <div className="px-6 py-3">
              <Text weight="strong">Namespace: {namespace}</Text>
            </div>

            <div className="flex flex-col">
              {Object.entries(namespaceDeployments).map(
                ([name, deployment]: [string, any]) => {
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
                    <Expand
                      key={name}
                      id={name}
                      className="border-t"
                      headerClassName="px-6"
                      heading={
                        <DeploymentHeading
                          healthPercentage={healthPercentage}
                          isHealthy={isHealthy}
                          name={name}
                          replicas={replicas}
                        />
                      }
                    >
                      <DeploymentDetails deployment={deployment} />
                    </Expand>
                  )
                }
              )}
            </div>
          </Card>
        ))
      ) : (
        <Card>
          <Text className="mx-auto" theme="neutral">
            No deployments
          </Text>
        </Card>
      )}
    </div>
  )
}

const DeploymentHeading = ({ healthPercentage, isHealthy, name, replicas }) => {
  return (
    <div className="flex items-center justify-between w-full">
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
}

const DeploymentDetails = ({ deployment }) => {
  const status = deployment.status
  const replicas = {
    desired: status?.replicas || 0,
    ready: status?.readyReplicas || 0,
    available: status?.availableReplicas || 0,
    updated: status?.updatedReplicas || 0,
  }

  const isHealthy =
    replicas.ready === replicas.desired &&
    replicas.available === replicas.desired

  return (
    <div className="bg-black/2 dark:bg-white/2 p-6 border-t flex flex-col gap-6">
      <div className="flex items-center gap-6">
        <LabeledValue label="Desired">
          <Text variant="base" weight="strong" theme="brand">
            {replicas?.desired}
          </Text>
        </LabeledValue>
        <LabeledValue label="Ready">
          <Text variant="base" weight="strong" theme="success">
            {replicas?.ready}
          </Text>
        </LabeledValue>
        <LabeledValue label="Available">
          <Text variant="base" weight="strong" theme="info">
            {replicas?.available}
          </Text>
        </LabeledValue>
        <LabeledValue label="Updated">
          <Text variant="base" weight="strong" theme="warn">
            {replicas?.updated}
          </Text>
        </LabeledValue>
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">Conditions</Text>
        <KeyValueList
          values={status?.conditions?.map((condition) => ({
            key: condition?.type,
            value: `${condition?.reason}: ${condition?.status}`,
            type: 'string',
          }))}
        />
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">Metadata</Text>

        <KeyValueList
          values={[
            {
              key: 'Created',
              value: new Date(
                deployment.metadata.creationTimestamp
              ).toLocaleString(),
              type: 'string',
            },
            {
              key: 'Generation',
              value: String(deployment.metadata.generation),
              type: 'number',
            },
            {
              key: 'Resource Version',
              value: deployment.metadata.resourceVersion,
              type: 'string',
            },
            {
              key: 'UID',
              value: deployment.metadata.uid,
              type: 'string',
            },
          ]}
        />
      </div>
    </div>
  )
}
