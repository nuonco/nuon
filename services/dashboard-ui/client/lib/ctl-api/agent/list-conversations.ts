import { api } from '@/lib/api'
import type { TConversation } from './create-conversation'

export const listConversations = ({ orgId }: { orgId: string }) =>
  api<TConversation[]>({
    path: `conversations`,
    pathVersion: '',
    baseUrl: `/api/orgs/${orgId}`,
  })
