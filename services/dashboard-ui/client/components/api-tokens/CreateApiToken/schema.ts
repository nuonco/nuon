import { z } from 'zod'

export const createApiTokenSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  role: z.string().min(1, 'Select a role'),
  duration: z.string().min(1, 'Select an expiry'),
})

export type CreateApiTokenValues = z.infer<typeof createApiTokenSchema>
