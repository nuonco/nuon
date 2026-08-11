import { z } from 'zod'
import type { Interests } from '@/components/interests'
import type { SubscriptionMatch } from '@/components/match/types'

export const createWebhookSchema = z.object({
  webhookUrl: z
    .string()
    .trim()
    .regex(/^https?:\/\/.+/i, 'Enter a valid http or https URL'),
  webhookSecret: z.string(),
  match: z.custom<SubscriptionMatch | undefined>(() => true),
  interests: z.custom<Interests>(() => true),
})

export type CreateWebhookValues = z.infer<typeof createWebhookSchema>
