import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Expand } from '@/components/common/Expand'
import { KeyValueList } from '@/components/common/KeyValueList'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'

export const IngressesDetails = ({ ingresses }) => {
  const hasIngresses = Object.keys(ingresses).length > 0
  const allIngresses = Object.values(ingresses).flatMap((namespace) =>
    Object.entries(namespace as any)
  )
  const publicCount = allIngresses.filter(
    ([_, ingress]: [string, any]) =>
      ingress.metadata?.annotations?.['alb.ingress.kubernetes.io/scheme'] ===
      'internet-facing'
  ).length
  const internalCount = allIngresses.length - publicCount
  const sslEnabledCount = allIngresses.filter(
    ([_, ingress]: [string, any]) =>
      ingress.metadata?.annotations?.[
        'alb.ingress.kubernetes.io/certificate-arn'
      ] ||
      ingress.metadata?.annotations?.[
        'alb.ingress.kubernetes.io/listen-ports'
      ]?.includes('HTTPS')
  ).length

  return (
    <div className="flex flex-col gap-2">
      <Text weight="strong">Ingresses details</Text>

      {hasIngresses ? (
        <>
          {/* <div className="flex items-center gap-10">
              <LabeledValue label="Total ingresses">
              <Text variant="base" weight="strong" theme="brand">
              {allIngresses?.length}
              </Text>
              </LabeledValue>
              <LabeledValue label="Public">
              <Text variant="base" weight="strong" theme="warn">
              {publicCount}
              </Text>
              </LabeledValue>
              <LabeledValue label="Internal">
              <Text variant="base" weight="strong" theme="info">
              {internalCount}
              </Text>
              </LabeledValue>

              <LabeledValue label="SSL Enabled">
              <Text variant="base" weight="strong" theme="success">
              {sslEnabledCount}
              </Text>
              </LabeledValue>
              </div> */}
          {Object.entries(ingresses).map(([namespace, namespaceIngresses]) => (
            <Card className="!p-0 !gap-0" key={namespace}>
              <div className="px-6 py-3">
                <Text weight="strong">Namespace: {namespace}</Text>
              </div>

              <div>
                {Object.entries(namespaceIngresses).map(([name, ingress]) => (
                  <Expand
                    id={name}
                    key={name}
                    className="border-t"
                    headerClassName="px-6"
                    heading={<IngressHeading name={name} ingress={ingress} />}
                  >
                    <IngressDetails ingress={ingress} />
                  </Expand>
                ))}
              </div>
            </Card>
          ))}
        </>
      ) : (
        <Card>
          <Text className="mx-auto" theme="neutral">
            No ingresses configured
          </Text>
        </Card>
      )}
    </div>
  )
}

const IngressHeading = ({ name, ingress }) => {
  const annotations = ingress.metadata?.annotations || {}
  const hostname = annotations['external-dns.alpha.kubernetes.io/hostname']
  const scheme = annotations['alb.ingress.kubernetes.io/scheme']
  const certificateArn =
    annotations['alb.ingress.kubernetes.io/certificate-arn']
  const listenPorts = annotations['alb.ingress.kubernetes.io/listen-ports']

  // Parse listen ports
  let ports = []
  try {
    if (listenPorts) {
      const parsed = JSON.parse(listenPorts)
      ports = parsed.map(
        (port: any) => Object.keys(port)[0] + ':' + Object.values(port)[0]
      )
    }
  } catch (e) {
    // Ignore parsing errors
  }

  const isPublic = scheme === 'internet-facing'
  const hasSSL = certificateArn || ports.some((p) => p.includes('HTTPS'))

  return (
    <div className="flex items-center justify-between w-full">
      <div className="flex items-center gap-2">
        <Text variant="body" weight="strong">
          {name}
        </Text>
        <Text variant="subtext" theme="neutral">
          {hostname}
        </Text>
      </div>
      <div className="flex items-center space-x-2">
        <Badge size="sm" theme={isPublic ? 'warn' : 'info'} variant="code">
          {isPublic ? 'PUBLIC' : 'INTERNAL'}
        </Badge>
        {hasSSL && (
          <Badge size="sm" theme="success" variant="code">
            SSL
          </Badge>
        )}
      </div>
    </div>
  )
}

const IngressDetails = ({ ingress }) => {
  const annotations = ingress.metadata?.annotations || {}
  const hostname = annotations['external-dns.alpha.kubernetes.io/hostname']
  const loadBalancerHostname =
    ingress.status?.loadBalancer?.ingress?.[0]?.hostname
  const scheme = annotations['alb.ingress.kubernetes.io/scheme']
  const targetType = annotations['alb.ingress.kubernetes.io/target-type']
  const certificateArn =
    annotations['alb.ingress.kubernetes.io/certificate-arn']
  const listenPorts = annotations['alb.ingress.kubernetes.io/listen-ports']

  // Parse listen ports
  let ports = []
  try {
    if (listenPorts) {
      const parsed = JSON.parse(listenPorts)
      ports = parsed.map(
        (port: any) => Object.keys(port)[0] + ':' + Object.values(port)[0]
      )
    }
  } catch (e) {
    // Ignore parsing errors
  }

  const isPublic = scheme === 'internet-facing'

  return (
    <div className="bg-black/2 dark:bg-white/2 p-6 border-t flex flex-col gap-6 max-h-full overflow-y-auto">
      <div className="flex items-start justify-between flew-wrap gap-6">
        <div className="flex flex-col gap-2">
          <Text weight="strong">Configuration</Text>
          <div className="flex items-center gap-8">
            <LabeledValue label="Scheme">
              <Badge
                variant="code"
                size="sm"
                theme={isPublic ? 'warn' : 'info'}
              >
                {scheme || 'Unknown'}
              </Badge>
            </LabeledValue>
            <LabeledValue label="Target">
              {targetType || 'Unknown'}
            </LabeledValue>
            {ports.length > 0 && (
              <LabeledValue label="Ports">
                <Text variant="subtext" family="mono">
                  {ports.join(', ')}
                </Text>
              </LabeledValue>
            )}
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <Text weight="strong">Endpoints</Text>
          <div className="flex items-center gap-8">
            <LabeledValue label="External">
              <Text
                className="flex items-center gap-2"
                variant="subtext"
                family="mono"
              >
                {hostname}{' '}
                <ClickToCopyButton textToCopy={loadBalancerHostname} />
              </Text>
            </LabeledValue>

            <LabeledValue label="Load balancer">
              <Text
                className="flex items-center gap-2"
                variant="subtext"
                family="mono"
              >
                {loadBalancerHostname}{' '}
                <ClickToCopyButton textToCopy={loadBalancerHostname} />
              </Text>
            </LabeledValue>
          </div>
        </div>
      </div>

      {certificateArn && (
        <div className="flex flex-col gap-2">
          <Text weight="strong">SSL Certificate</Text>
          <div className="flex items-center gap-2">
            <Badge variant="code" size="lg" theme="success">
              {certificateArn}
            </Badge>
            <ClickToCopyButton textToCopy={certificateArn} />
          </div>
        </div>
      )}

      {annotations['alb.ingress.kubernetes.io/healthcheck-path'] && (
        <div className="flex flex-col gap-2">
          <Text weight="strong">Health check</Text>

          <div className="flex items-center gap-8">
            <LabeledValue label="Path">
              <Text variant="subtext" family="mono">
                {annotations['alb.ingress.kubernetes.io/healthcheck-path']}
              </Text>
            </LabeledValue>

            <LabeledValue label="Timeout">
              <Text variant="subtext" family="mono">
                {`${
                  annotations[
                    'alb.ingress.kubernetes.io/healthcheck-timeout-seconds'
                  ]
                }s`}
              </Text>
            </LabeledValue>

            <LabeledValue label="Threshold">
              <Text variant="subtext" family="mono">
                {`${
                  annotations[
                    'alb.ingress.kubernetes.io/healthy-threshold-count'
                  ]
                }/${
                  annotations[
                    'alb.ingress.kubernetes.io/unhealthy-threshold-count'
                  ]
                }`}
              </Text>
            </LabeledValue>
          </div>
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Text weight="strong">Metadata</Text>
        <KeyValueList
          values={[
            {
              key: 'Created',
              value: new Date(
                ingress.metadata.creationTimestamp
              ).toLocaleString(),
              type: 'string',
            },
            {
              key: 'Generation',
              value: String(ingress.metadata.generation),
              type: 'number',
            },
            {
              key: 'Resource Version',
              value: ingress.metadata.resourceVersion,
              type: 'string',
            },
            { key: 'UID', value: ingress.metadata.uid, type: 'string' },
          ]}
        />
      </div>

      <div className="flex flex-col gap-2">
        <Text weight="strong">AWS load balancer annotations</Text>
        <KeyValueList
          values={Object.entries(annotations)
            .filter(([key]) => key.startsWith('alb.ingress.kubernetes.io/'))
            .map(([key, value]) => ({
              key: key.replace('alb.ingress.kubernetes.io/', ''),
              value: String(value),
              type: 'string',
            }))}
        />
      </div>
    </div>
  )
}
