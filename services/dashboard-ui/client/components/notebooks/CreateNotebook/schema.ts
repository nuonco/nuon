import { z } from 'zod'

export const createNotebookSchema = z.object({
  name: z.string().trim().min(1, 'Name is required').max(255),
  description: z.string().max(2000).optional(),
})

export type CreateNotebookValues = z.infer<typeof createNotebookSchema>
