import { z } from 'zod'

export const accountSchema = z.object({
  name: z.string().trim().min(1, 'Connection name is required'),
  target_id: z.string().regex(/^\d{12}$/, 'Enter a 12-digit AWS account ID'),
  role_name: z
    .string()
    .min(1, 'Role name is required')
    .max(64, 'Role name must be at most 64 characters')
    .regex(/^[\w+=,.@-]+$/, 'Use letters, numbers, or _+=,.@-'),
})
export const accessSchema = z.object({ preset: z.enum(['stacks', 'custom']) })
export type TAccountValues = z.infer<typeof accountSchema>
