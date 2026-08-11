import { z } from 'zod'

export const updateRunnerSchema = z.object({
  tag: z.string().trim().min(1, 'Enter a value to update to'),
})

export type UpdateRunnerValues = z.infer<typeof updateRunnerSchema>
