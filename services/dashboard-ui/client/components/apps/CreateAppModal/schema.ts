import { z } from 'zod'

export const createAppSchema = z.object({
  name: z.string().trim().min(1, 'Name is required').max(255),
})

export type CreateAppValues = z.infer<typeof createAppSchema>
