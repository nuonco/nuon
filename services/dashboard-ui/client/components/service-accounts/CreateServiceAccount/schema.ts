import { z } from 'zod'

export const createServiceAccountSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  role: z.string().min(1, 'Select a role'),
})

export type CreateServiceAccountValues = z.infer<
  typeof createServiceAccountSchema
>
