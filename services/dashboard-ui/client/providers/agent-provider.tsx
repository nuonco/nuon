import {
  createContext,
  useCallback,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useMutation } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { createConversation } from '@/lib/ctl-api/agent'

export type TAgentMessage = {
  id: string
  role: 'user' | 'assistant'
  content: string
  toolCalls?: TAgentToolCall[]
  blocks?: TAgentBlock[]
  isStreaming?: boolean
}

export type TAgentBlock =
  | { type: 'text'; content: string }
  | { type: 'tool_call'; toolCall: TAgentToolCall }

export type TAgentToolCall = {
  id: string
  tool: string
  args?: string
  result?: string
  status: 'running' | 'complete' | 'error'
}

type TAgentStatus = 'idle' | 'thinking' | 'acting' | 'done' | 'error'

export type TAgentContext = {
  isOpen: boolean
  setIsOpen: (open: boolean) => void
  messages: TAgentMessage[]
  status: TAgentStatus
  sendMessage: (content: string) => void
  isSending: boolean
  conversationId: string | null
  clearConversation: () => void
}

export const AgentContext = createContext<TAgentContext | undefined>(
  undefined
)

export function AgentProvider({ children }: { children: ReactNode }) {
  const { org } = useOrg()
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState<TAgentMessage[]>([])
  const [status, setStatus] = useState<TAgentStatus>('idle')
  const [isSending, setIsSending] = useState(false)
  const [conversationId, setConversationId] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)
  const assistantMsgIdRef = useRef(0)

  const { mutateAsync: createConv } = useMutation({
    mutationFn: () => createConversation({ orgId: org!.id }),
  })

  const clearConversation = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setMessages([])
    setConversationId(null)
    setStatus('idle')
    setIsSending(false)
  }, [])

  const sendMessage = useCallback(
    async (content: string) => {
      if (!org?.id || isSending) return

      setIsSending(true)
      setStatus('thinking')

      const userMsg: TAgentMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content,
      }
      setMessages((prev) => [...prev, userMsg])

      try {
        let convId = conversationId
        if (!convId) {
          const conv = await createConv()
          convId = conv.id
          setConversationId(convId)
        }

        const assistantId = `assistant-${++assistantMsgIdRef.current}`
        const toolCalls = new Map<string, TAgentToolCall>()
        const blocks: TAgentBlock[] = []

        setMessages((prev) => [
          ...prev,
          {
            id: assistantId,
            role: 'assistant',
            content: '',
            isStreaming: true,
            toolCalls: [],
          },
        ])

        const updateAssistant = (
          updater: (msg: TAgentMessage) => TAgentMessage
        ) => {
          setMessages((prev) =>
            prev.map((m) => (m.id === assistantId ? updater(m) : m))
          )
        }

        const response = await fetch(
          `/api/orgs/${org.id}/conversations/${convId}/messages`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ content }),
          }
        )

        if (!response.ok || !response.body) {
          throw new Error('Failed to send message')
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() ?? ''

          let currentEvent = ''
          for (const line of lines) {
            if (line.startsWith('event: ')) {
              currentEvent = line.slice(7)
            } else if (line.startsWith('data: ') && currentEvent) {
              try {
                const data = JSON.parse(line.slice(6))
                handleSSEEvent(
                  currentEvent,
                  data,
                  updateAssistant,
                  toolCalls,
                  blocks,
                  setStatus
                )
              } catch {
                // skip malformed events
              }
              currentEvent = ''
            }
          }
        }

        updateAssistant((m) => ({ ...m, isStreaming: false }))
        setStatus('idle')
      } catch {
        setStatus('error')
      } finally {
        setIsSending(false)
      }
    },
    [org?.id, conversationId, isSending, createConv]
  )

  return (
    <AgentContext.Provider
      value={{
        isOpen,
        setIsOpen,
        messages,
        status,
        sendMessage,
        isSending,
        conversationId,
        clearConversation,
      }}
    >
      {children}
    </AgentContext.Provider>
  )
}

function handleSSEEvent(
  event: string,
  data: any,
  updateAssistant: (
    updater: (msg: TAgentMessage) => TAgentMessage
  ) => void,
  toolCalls: Map<string, TAgentToolCall>,
  blocks: TAgentBlock[],
  setStatus: (s: TAgentStatus) => void
) {
  const syncBlocks = () => {
    const snapshot = blocks.map((b) =>
      b.type === 'tool_call'
        ? { ...b, toolCall: { ...b.toolCall } }
        : { ...b }
    )
    updateAssistant((m) => ({ ...m, blocks: snapshot }))
  }

  switch (event) {
    case 'text': {
      const text = data?.content ?? ''
      const last = blocks[blocks.length - 1]
      if (last?.type === 'text') {
        last.content += text
      } else {
        blocks.push({ type: 'text', content: text })
      }
      updateAssistant((m) => ({
        ...m,
        content: m.content + text,
        blocks: blocks.map((b) =>
          b.type === 'tool_call'
            ? { ...b, toolCall: { ...b.toolCall } }
            : { ...b }
        ),
      }))
      break
    }

    case 'tool_call':
      if (data?.id) {
        const tc: TAgentToolCall = {
          id: data.id,
          tool: data.tool,
          status: 'running',
        }
        toolCalls.set(data.id, tc)
        blocks.push({ type: 'tool_call', toolCall: tc })
        syncBlocks()
      }
      break

    case 'tool_result':
      if (data?.id && toolCalls.has(data.id)) {
        const tc = toolCalls.get(data.id)!
        tc.result = data.result
        tc.status = data.status === 'error' ? 'error' : 'complete'
        syncBlocks()
      }
      break

    case 'status':
      if (data?.status) {
        setStatus(data.status as TAgentStatus)
      }
      break

    case 'error':
      blocks.push({
        type: 'text',
        content: `Error: ${data?.message ?? 'Unknown error'}`,
      })
      updateAssistant((m) => ({
        ...m,
        content:
          m.content +
          (m.content ? '\n\n' : '') +
          `Error: ${data?.message ?? 'Unknown error'}`,
        blocks: blocks.map((b) =>
          b.type === 'tool_call'
            ? { ...b, toolCall: { ...b.toolCall } }
            : { ...b }
        ),
        isStreaming: false,
      }))
      setStatus('error')
      break
  }
}
