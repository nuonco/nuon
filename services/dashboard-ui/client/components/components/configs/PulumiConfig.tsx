import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TKeyValue } from '@/types'

interface PulumiConfigModalProps extends IModal {
  config: Record<string, string>
}

interface PulumiEnvVarsModalProps extends IModal {
  envVars: Record<string, string>
}

export const PulumiConfigModal = ({
  config,
  ...props
}: PulumiConfigModalProps) => {
  const keyValuePairs: TKeyValue[] = Object.entries(config).map(([key, value]) => ({
    key,
    value,
    type: 'string',
  }))

  const configText = Object.entries(config)
    .map(([key, value]) => `${key}: "${value}"`)
    .join('\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" flex className="gap-2">
          <Icon variant="List" size="20" />
          Pulumi config
        </Text>
      }
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex justify-between items-center gap-4">
          <Text variant="body">
            This is the Pulumi config for this component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={configText} className="w-fit" />
          </div>
        </div>
        <KeyValueList values={keyValuePairs} />
      </div>
    </Modal>
  )
}

export const PulumiEnvVarsModal = ({
  envVars,
  ...props
}: PulumiEnvVarsModalProps) => {
  const keyValuePairs: TKeyValue[] = Object.entries(envVars).map(([key, value]) => ({
    key,
    value,
    type: 'string',
  }))

  const envVarsText = Object.entries(envVars)
    .map(([key, value]) => `${key}="${value}"`)
    .join('\n')

  return (
    <Modal
      heading={
        <Text variant="h3" weight="strong" flex className="gap-2">
          <Icon variant="List" size="20" />
          Pulumi env vars
        </Text>
      }
      size="lg"
      className="!max-h-[80vh]"
      childrenClassName="overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex justify-between items-center gap-4">
          <Text variant="body">
            This is the environment variables for this Pulumi component configuration.
          </Text>
          <div className="flex justify-end">
            <ClickToCopyButton textToCopy={envVarsText} className="w-fit" />
          </div>
        </div>
        <KeyValueList values={keyValuePairs} />
      </div>
    </Modal>
  )
}
