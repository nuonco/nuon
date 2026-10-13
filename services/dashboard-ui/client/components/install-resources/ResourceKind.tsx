import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TInstallResource } from '@/types'

const kindDescriptions: Record<string, string> = {
  Ingress: 'Routes external HTTP and HTTPS traffic to Services.',
  Service: 'Provides a stable network endpoint for a selected set of Pods.',
  Deployment: 'Manages ReplicaSets and rolls out application updates.',
  ReplicaSet: 'Maintains the desired number of matching Pods.',
  Pod: 'Runs one or more containers together.',
  StatefulSet: 'Manages Pods with stable identities and persistent storage.',
  DaemonSet: 'Runs a Pod on each eligible cluster node.',
  Job: 'Runs Pods to completion.',
  CronJob: 'Creates Jobs on a schedule.',
  HorizontalPodAutoscaler:
    'Scales a workload to meet its configured metric targets.',
  Gateway: 'Configures listeners for incoming traffic.',
  HTTPRoute:
    'Attaches HTTP routing rules to a Gateway and references backends.',
  ConfigMap: 'Provides configuration to workloads; it has no health verdict.',
  Secret: 'Provides sensitive configuration. Only references are shown here.',
  HTTPProbe: 'Checks an HTTP endpoint from the runner.',
  TCPProbe: 'Checks whether the runner can establish a TCP connection.',
  ExecProbe: 'Runs a configured health check command from the runner.',
  CustomCheck: 'A health verdict pushed by an external check.',
  PersistentVolumeClaim: 'Requests persistent storage for a workload.',
}

export const ResourceKind = ({
  resource,
  compact = false,
}: {
  resource: TInstallResource
  compact?: boolean
}) => (
  <Tooltip
    tipContent={
      resource.provider === 'kubernetes' &&
      ['HTTPProbe', 'TCPProbe', 'ExecProbe', 'CustomCheck'].includes(
        resource.kind || ''
      )
        ? resource.kind
        : kindDescriptions[resource.kind || ''] || resource.kind
    }
  >
    <span className="inline-flex items-center gap-2 whitespace-nowrap">
      {resource.provider === 'kubernetes' ? (
        <Icon variant="Kubernetes" theme="info" size={20} />
      ) : null}
      <Text variant="subtext" family="mono">
        {compact && resource.kind === 'HorizontalPodAutoscaler'
          ? 'HPA'
          : compact && resource.kind === 'PersistentVolumeClaim'
            ? 'PVC'
            : resource.kind}
      </Text>
    </span>
  </Tooltip>
)
