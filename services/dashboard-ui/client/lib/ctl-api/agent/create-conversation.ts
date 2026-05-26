import { api } from '@/lib/api'

export type TConversation = {
  id: string
  org_id: string
  messages: TConversationMessage[]
  created_at: string
  updated_at: string
}

export type TConversationMessage = {
  role: string
  content?: string
  tool_calls?: TToolCall[]
  tool_result?: TToolResult
}

export type TToolCall = {
  id: string
  name: string
  args: string
}

export type TToolResult = {
  tool_call_id: string
  content: string
  is_error?: boolean
}

export const createConversation = ({ orgId }: { orgId: string }) =>
  api<TConversation>({
    path: `conversations`,
    pathVersion: '',
    baseUrl: `/api/orgs/${orgId}`,
    method: 'POST',
  })
