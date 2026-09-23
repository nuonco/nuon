import { z } from 'zod'

export const previewConfigSchema = z.object({
  mode: z.enum(['none', 'plan-only', 'apply', 'build-only']),
  installId: z.string(),
  setStatuses: z.boolean(),
  comment: z.boolean(),
  ignoreDrafts: z.boolean(),
})

export type PreviewConfigFormValues = z.infer<typeof previewConfigSchema>
