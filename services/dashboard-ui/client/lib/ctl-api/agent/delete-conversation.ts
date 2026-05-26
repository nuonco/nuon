import { api } from '@/lib/api'

export const deleteConversation = ({
  orgId,
  conversationId,
}: {
  orgId: string
  conversationId: string
}) =>
  api<void>({
    path: `conversations/${conversationId}`,
    pathVersion: '',
    baseUrl: `/api/orgs/${orgId}`,
    method: 'DELETE',
  })
