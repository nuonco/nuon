import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useAgent } from '@/hooks/use-agent'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { MessageThread } from './MessageThread'
import { MessageInput } from './MessageInput'

const statusLabels: Record<string, string> = {
  thinking: 'Thinking',
  acting: 'Running tools',
  error: 'Error',
}

export function AgentPanel(props: IPanel) {
  const { messages, status, sendMessage, isSending } = useAgent()

  const heading = (
    <div className="flex items-center gap-2">
      <Icon variant="SparkleIcon" size={16} className="text-primary-600" />
      <Text variant="base" weight="strong">Agent</Text>
      {status !== 'idle' && status !== 'done' && (
        <Badge size="sm" theme={status === 'error' ? 'error' : 'info'}>
          {statusLabels[status]}
        </Badge>
      )}
    </div>
  )

  return (
    <Panel
      heading={heading}
      childrenClassName="!p-0 !gap-0 overflow-hidden"
      className="!overflow-hidden"
      onClose={() => props.onClose?.()}
      size="half"
      {...props}
    >
      <div className="flex flex-col flex-auto overflow-hidden">
        <MessageThread messages={messages} />
        <MessageInput onSend={sendMessage} disabled={isSending} />
      </div>
    </Panel>
  )
}
