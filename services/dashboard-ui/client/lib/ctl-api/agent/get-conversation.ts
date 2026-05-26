import { api } from '@/lib/api'
import type { TConversation } from './create-conversation'

export const getConversation = ({
  orgId,
  conversationId,
}: {
  orgId: string
  conversationId: string
}) =>
  api<TConversation>({
    path: `conversations/${conversationId}`,
    pathVersion: '',
    baseUrl: `/api/orgs/${orgId}`,
  })
