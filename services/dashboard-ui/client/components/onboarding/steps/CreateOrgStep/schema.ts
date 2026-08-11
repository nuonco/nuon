import { z } from 'zod'

export const createOrgSchema = z.object({
  orgName: z.string().trim().min(1, 'Organization name is required'),
})

export type CreateOrgValues = z.infer<typeof createOrgSchema>
