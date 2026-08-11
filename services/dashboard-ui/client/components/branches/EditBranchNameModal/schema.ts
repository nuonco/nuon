import { z } from 'zod'

export const editBranchSchema = z.object({
  branchName: z.string().trim().min(1, 'Branch name cannot be empty'),
  useVcs: z.boolean(),
  directory: z.string(),
  pathFilter: z.string(),
})

export type EditBranchValues = z.infer<typeof editBranchSchema>
