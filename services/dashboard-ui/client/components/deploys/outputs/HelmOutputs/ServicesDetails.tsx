import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'

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

        <div className="flex flex-wrap items-start gap-x-16 gap-y-4">
          <LabeledValue label="Created">
            <Time
              variant="subtext"
              time={service.metadata.creationTimestamp}
              format="short-datetime"
            />
          </LabeledValue>
          <LabeledValue label="UID">{service.metadata.uid}</LabeledValue>
          <LabeledValue label="Resource version">
            {service.metadata.resourceVersion}
          </LabeledValue>
          <LabeledValue label="Type">
            {service.spec?.type || 'Unknown'}
          </LabeledValue>
          <LabeledValue label="Cluster IP">
            {service.spec?.clusterIP || 'None'}
          </LabeledValue>
          <LabeledValue label="Session affinity">
            {service.spec?.sessionAffinity || 'None'}
          </LabeledValue>
        </div>
      </div>
    </div>
  )
}
