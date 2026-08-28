import { Button, type IButtonAsButton } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Divider } from '@/components/common/Divider'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import type { TInstallResource } from '@/types'

function parseResourceDetails(details?: string): unknown | undefined {
  if (!details) return undefined
  try {
    return JSON.parse(details)
  } catch {
    return undefined
  }
}

export interface IInstallResourceDetailPanel extends IPanel {
  installResource: TInstallResource
}

export const InstallResourceDetailPanel = ({
  installResource: resource,
  ...props
}: IInstallResourceDetailPanel) => {
  const parsedDetails = parseResourceDetails(resource?.details)

  return (
    <Panel heading={resource?.name || 'Resource'} size="half" {...props}>
      <div className="flex items-center gap-2">
        <Status variant="badge" status={resource?.health} />
        {resource?.kind ? <Text variant="subtext">{resource.kind}</Text> : null}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <LabeledValue label="Provider">
          {resource?.provider || <Icon variant="MinusIcon" />}
        </LabeledValue>
        <LabeledValue label="Observed">
          {resource?.observed_at ? (
            <Time
              variant="subtext"
              time={resource.observed_at}
              format="relative"
            />
          ) : (
            <Icon variant="MinusIcon" />
          )}
        </LabeledValue>
        {/* Kubernetes-shaped fields are omitted rather than shown empty: a probe
            or custom check has no namespace, api group, or native status, and a
            column of dashes reads as missing data rather than not applicable. */}
        {resource?.namespace ? (
          <LabeledValue label="Namespace">{resource.namespace}</LabeledValue>
        ) : null}
        {resource?.api_group ? (
          <LabeledValue label="API group">{resource.api_group}</LabeledValue>
        ) : null}
        {resource?.native_status ? (
          <LabeledValue label="Native status">
            {resource.native_status}
          </LabeledValue>
        ) : null}
        {resource?.runner_id ? (
          <LabeledValue label="Runner ID">
            <ID>{resource.runner_id}</ID>
          </LabeledValue>
        ) : null}
      </div>

      {resource?.message ? (
        <LabeledValue label="Message">{resource.message}</LabeledValue>
      ) : null}

      <Divider dividerWord="Details" />

      {parsedDetails !== undefined ? (
        <JSONViewer data={parsedDetails} expanded={2} showCopy />
      ) : resource?.details ? (
        <CodeBlock language="text" showCopy>
          {resource.details}
        </CodeBlock>
      ) : (
        <Text variant="subtext" theme="neutral">
          No details reported for this resource.
        </Text>
      )}
    </Panel>
  )
}

export interface IInstallResourceDetailPanelButton extends IButtonAsButton {
  onOpen: () => void
}

export const InstallResourceDetailPanelButton = ({
  onOpen,
  ...props
}: IInstallResourceDetailPanelButton) => {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={onOpen}
      aria-label="View resource details"
      {...props}
    >
      Details
    </Button>
  )
}
