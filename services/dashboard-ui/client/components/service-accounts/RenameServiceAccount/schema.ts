import { z } from 'zod'

export const renameServiceAccountSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
})

export type RenameServiceAccountValues = z.infer<
  typeof renameServiceAccountSchema
>
