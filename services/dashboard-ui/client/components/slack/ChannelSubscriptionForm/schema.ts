import { z } from 'zod'
import type { Interests } from '@/components/interests'
import type { SubscriptionMatch } from '@/components/match/types'

export type ChannelSubscriptionMode = 'create' | 'edit'

export interface ChannelSubscriptionValues {
  channelId: string
  channelName: string
  match: SubscriptionMatch | undefined
  interests: Interests
}

export type ChannelSubscriptionOutput = {
  orgLinkId: string
  channelId: string
  channelName: string
  match: SubscriptionMatch | undefined
  interests: Interests
}

export const buildChannelSubscriptionSchema = (mode: ChannelSubscriptionMode) =>
  z.object({
    channelId:
      mode === 'create' ? z.string().min(1, 'Select a channel') : z.string(),
    channelName: z.string(),
    match: z.custom<SubscriptionMatch | undefined>(() => true),
    interests: z.custom<Interests>(() => true),
  })
