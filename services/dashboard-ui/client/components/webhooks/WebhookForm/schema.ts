import { z } from 'zod'
import type { Interests } from '@/components/interests'
import type { SubscriptionMatch } from '@/components/match/types'

export type WebhookFormMode = 'create' | 'edit'
export type WebhookSecretMode = 'keep' | 'rotate' | 'clear'

export interface WebhookFormValues {
  webhookUrl: string
  secretMode: WebhookSecretMode
  webhookSecret: string
  match: SubscriptionMatch | undefined
  interests: Interests
}

export type WebhookFormOutput = {
  webhookUrl: string
  webhookSecret: string | undefined
  match: SubscriptionMatch | undefined
  interests: Interests
}

export const buildWebhookSchema = (mode: WebhookFormMode) =>
  z
    .object({
      webhookUrl:
        mode === 'create'
          ? z
              .string()
              .trim()
              .regex(/^https?:\/\/.+/i, 'Enter a valid http or https URL')
          : z.string(),
      secretMode: z.enum(['keep', 'rotate', 'clear']),
      webhookSecret: z.string(),
      match: z.custom<SubscriptionMatch | undefined>(() => true),
      interests: z.custom<Interests>(() => true),
    })
    .superRefine((value, ctx) => {
      if (
        mode === 'edit' &&
        value.secretMode === 'rotate' &&
        value.webhookSecret.trim().length === 0
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['webhookSecret'],
          message: 'Enter a new signing secret',
        })
      }
    })
