import { z } from 'zod'
import type { TVCSConnectionRepo } from '@/types'

export type BranchFormMode = 'create' | 'edit'

export const branchFormSchema = z.object({
  name: z.string().trim().min(1, 'Branch name is required'),
  useVcs: z.boolean(),
  directory: z.string(),
  pathFilter: z.string(),
  ignoreAllChanges: z.boolean(),
})

export type BranchFormValues = z.infer<typeof branchFormSchema>

export interface BranchFormOutput {
  name: string
  useVcs: boolean
  selectedVcsConnectionId: string
  selectedRepo: TVCSConnectionRepo | null
  selectedBranch: string
  directory: string
  pathFilter: string
  ignoreAllChanges: boolean
}
