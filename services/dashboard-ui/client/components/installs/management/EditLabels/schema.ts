import { z } from 'zod'

export const editLabelsSchema = z.object({
  labels: z.array(
    z.object({
      key: z.string().trim().min(1, 'Key is required'),
      value: z.string(),
    })
  ),
})

export type EditLabelsValues = z.infer<typeof editLabelsSchema>
