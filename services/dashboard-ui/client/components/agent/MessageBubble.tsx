import type { TAgentMessage } from '@/providers/agent-provider'
import { Text } from '@/components/common/Text'
import { Markdown } from '@/components/common/Markdown/Markdown'

interface IMessageBubble {
  message: TAgentMessage
}

export function MessageBubble({ message }: IMessageBubble) {
  const isUser = message.role === 'user'

  if (isUser) {
    return (
      <div className="flex justify-end">
        <div className="max-w-[85%] rounded-md bg-primary-600 px-3 py-2">
          <Text variant="body" className="whitespace-pre-wrap break-words !text-white">
            {message.content}
          </Text>
        </div>
      </div>
    )
  }

  if (!message.content && message.isStreaming && !message.toolCalls?.length) {
    return (
      <div className="flex items-center gap-1.5 py-1">
        <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 dark:bg-cool-grey-500" />
        <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 [animation-delay:150ms] dark:bg-cool-grey-500" />
        <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 [animation-delay:300ms] dark:bg-cool-grey-500" />
      </div>
    )
  }

  if (!message.content) return null

  return (
    <div className="flex justify-start">
      <div className="max-w-[85%] rounded-md border border-cool-grey-300 bg-cool-grey-50 px-3 py-2 dark:border-dark-grey-300 dark:bg-dark-grey-400">
        <div className="break-words text-sm [&_pre]:overflow-x-auto [&_pre]:max-w-full [&_table]:text-xs">                    
          <Markdown content={message.content} />
        </div>
      </div>
    </div>
  )
}
