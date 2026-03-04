import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { objectToKeyValueArray } from '@/utils/data-utils'

export const ServicesDetails = ({
  services,
}: {
  services: Record<string, any>
}) => {
  const hasServices = Object.keys(services).length > 0
  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Services details</Text>
      {hasServices ? (
        Object.entries(services).map(([namespace, namespaceServices]) => (
          <Card key={namespace} className="!p-0 !gap-0">
            <div className="px-6 py-3">
              <Text weight="strong">Namespace: {namespace}</Text>
            </div>

            <div className="flex flex-col">
              {Object.entries(namespaceServices).map(
                ([name, service]: [string, any]) => {
                  const isHealthy = true

                  return (
                    <Expand
                      key={name}
                      id={name}
                      className="border-t"
                      headerClassName="px-6"
                      heading={
                        <ServiceHeading isHealthy={isHealthy} name={name} />
                      }
                    >
                      <ServiceDetails service={service} />
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
            No services
          </Text>
        </Card>
      )}
    </div>
  )
}

const ServiceHeading = ({ isHealthy, name }) => {
  return (
    <div key={name} className="flex items-center justify-between w-full">
      <div className="flex items-center gap-2">
        <Status
          status={isHealthy ? 'healthy' : 'info'}
          isWithoutText
          variant="timeline"
        />
        <Text variant="body" weight="strong">
          {name}
        </Text>
      </div>
    </div>
  )
}

const ServiceDetails = ({ service }) => {
  return (
    <div className="bg-black/2 dark:bg-white/2 p-6 border-t flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Text weight="strong">Metadata</Text>

        <KeyValueList
          values={objectToKeyValueArray({
            Created: new Date(
              service.metadata.creationTimestamp
            ).toLocaleString(),
            UID: service.metadata.uid,
            'Resource Version': service.metadata.resourceVersion,
            Type: service.spec?.type || 'Unknown',
            'Cluster IP': service.spec?.clusterIP || 'None',
            'Session Affinity': service.spec?.sessionAffinity || 'None',
          })}
        />
      </div>
    </div>
  )
}
