import { z } from 'zod'

export const updateOrgNameSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
})

export type UpdateOrgNameValues = z.infer<typeof updateOrgNameSchema>
