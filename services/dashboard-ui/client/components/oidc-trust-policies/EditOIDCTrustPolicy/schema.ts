import { z } from 'zod'

export const editOIDCSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  issuerUrl: z
    .string()
    .trim()
    .regex(/^https?:\/\/.+/i, 'Must be an absolute http or https URL'),
  audience: z.string().trim().min(1, 'Audience is required'),
  role: z.string(),
  tokenDurationSeconds: z.string(),
  enabled: z.boolean(),
  claimConditions: z.array(z.object({ key: z.string(), value: z.string() })),
})

export type EditOIDCFormValues = z.infer<typeof editOIDCSchema>
