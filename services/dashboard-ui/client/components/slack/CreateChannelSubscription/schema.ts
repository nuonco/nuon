import { z } from 'zod'
import type { Interests } from '@/components/interests'
import type { SubscriptionMatch } from '@/components/match/types'

export const channelSubscriptionSchema = z.object({
  channelId: z.string().min(1, 'Select a channel'),
  channelName: z.string(),
  match: z.custom<SubscriptionMatch | undefined>(() => true),
  interests: z.custom<Interests>(() => true),
})

export type ChannelSubscriptionValues = z.infer<typeof channelSubscriptionSchema>
