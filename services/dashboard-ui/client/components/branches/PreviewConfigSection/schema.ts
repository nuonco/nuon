import { z } from 'zod'

export const previewConfigSchema = z.object({
  mode: z.enum(['plan-only', 'apply', 'build-only', 'disabled']),
  installId: z.string(),
  setStatuses: z.boolean(),
  comment: z.boolean(),
  ignoreDrafts: z.boolean(),
})

export type PreviewConfigFormValues = z.infer<typeof previewConfigSchema>
