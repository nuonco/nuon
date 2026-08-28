import { z } from 'zod'

export const previewConfigSchema = z.object({
  mode: z.enum(['plan-only', 'plan-infra', 'apply', 'build-only']),
  installId: z.string(),
  setStatuses: z.boolean(),
  comment: z.boolean(),
})

export type PreviewConfigFormValues = z.infer<typeof previewConfigSchema>
