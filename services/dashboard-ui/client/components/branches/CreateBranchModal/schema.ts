import { z } from 'zod'

export const createBranchSchema = z.object({
  name: z.string().trim().min(1, 'Branch name is required'),
  useVcs: z.boolean(),
  directory: z.string(),
  pathFilter: z.string(),
})

export type CreateBranchValues = z.infer<typeof createBranchSchema>
