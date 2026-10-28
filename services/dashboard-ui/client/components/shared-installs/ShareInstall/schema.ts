import { z } from 'zod'

export const shareInstallSchema = z
  .object({
    email: z.string().trim().email('Enter a valid email address'),
    approvals: z.boolean(),
    readme: z.boolean(),
    operations: z.boolean(),
    resources: z.boolean(),
  })
  .refine(
    (value) =>
      value.approvals || value.readme || value.operations || value.resources,
    {
      message: 'Select at least one thing to share',
      path: ['approvals'],
    }
  )

export type ShareInstallFormValues = z.infer<typeof shareInstallSchema>
