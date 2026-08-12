import { z } from 'zod'
import type { TPermissionEntry, TRoleContext } from '@/types'

export const roleFormSchema = z.object({
  title: z.string().trim().min(1, 'Name is required'),
  description: z.string(),
  contexts: z.custom<TRoleContext[]>(() => true),
  permissions: z.custom<TPermissionEntry[]>(() => true),
})

export type RoleFormValues = z.infer<typeof roleFormSchema>
