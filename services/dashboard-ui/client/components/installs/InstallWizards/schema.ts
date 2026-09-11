import { z } from 'zod'

export const installSetupSchema = z
  .object({
    installName: z.string().trim().min(1, 'Enter an install name'),
    region: z.string().trim().min(1, 'Enter an AWS region'),
    stackOwnership: z.enum(['customer', 'nuon']),
    roleArn: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.stackOwnership === 'nuon' && !values.roleArn.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Enter the IAM role ARN',
        path: ['roleArn'],
      })
    }
  })

export type TInstallSetupValues = z.infer<typeof installSetupSchema>
