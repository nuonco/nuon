import { useEffect, useRef, type ReactNode } from 'react'
import type { TAgentMessage } from '@/providers/agent-provider'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Markdown } from '@/components/common/Markdown/Markdown'
import { ToolCallCard } from './ToolCallCard'

interface IMessageThread {
  messages: TAgentMessage[]
}

export function MessageThread({ messages }: IMessageThread) {
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-3 px-4 md:px-6 text-center">
        <Icon variant="SparkleIcon" size={24} className="text-primary-500" />
        <Text variant="body" weight="strong">How can I help?</Text>
        <Text variant="subtext" theme="neutral" className="max-w-[240px]">
          I can help you configure apps, create components, set up installs, and
          diagnose issues.
        </Text>
      </div>
    )
  }

  const items: ReactNode[] = []
  for (const msg of messages) {
    if (msg.role === 'user') {
      items.push(
        <div key={msg.id} className="flex justify-end">
          <div className="max-w-[85%] rounded-md bg-primary-600 px-3 py-2">
            <Text variant="body" className="whitespace-pre-wrap break-words !text-white">
              {msg.content}
            </Text>
          </div>
        </div>
      )
      continue
    }

    if (msg.blocks?.length) {
      for (let i = 0; i < msg.blocks.length; i++) {
        const block = msg.blocks[i]
        if (block.type === 'tool_call') {
          items.push(
            <ToolCallCard key={block.toolCall.id} toolCall={block.toolCall} />
          )
        } else if (block.type === 'text' && block.content) {
          items.push(
            <div key={`${msg.id}-text-${i}`} className="flex justify-start">
              <div className="max-w-[85%] rounded-md border border-cool-grey-300 bg-cool-grey-50 px-3 py-2 dark:border-dark-grey-300 dark:bg-dark-grey-600">
                <div className="break-words text-sm [&_pre]:overflow-x-auto [&_pre]:max-w-full [&_table]:text-xs">
                  <Markdown content={block.content} />
                </div>
              </div>
            </div>
          )
        }
      }
    } else if (msg.isStreaming && !msg.content && !msg.toolCalls?.length) {
      items.push(
        <div key={msg.id} className="flex items-center gap-1.5 py-1">
          <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 dark:bg-cool-grey-500" />
          <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 [animation-delay:150ms] dark:bg-cool-grey-500" />
          <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-cool-grey-400 [animation-delay:300ms] dark:bg-cool-grey-500" />
        </div>
      )
    }
  }

  return (
    <div className="flex flex-1 flex-col gap-4 overflow-y-auto px-4 md:px-6 py-4">
      {items}
      <div ref={endRef} />
    </div>
  )
}
