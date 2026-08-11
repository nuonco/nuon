import { z } from 'zod'

export const changeServiceAccountRoleSchema = z.object({
  role: z.string().min(1, 'Select a role'),
})

export type ChangeServiceAccountRoleValues = z.infer<
  typeof changeServiceAccountRoleSchema
>
