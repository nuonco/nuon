import { z } from 'zod'

export const createCloudConnectionSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  targetId: z
    .string()
    .trim()
    .regex(/^\d{12}$/, 'Enter a 12-digit AWS account ID'),
  principal: z
    .string()
    .trim()
    .regex(/^arn:aws:iam::\d{12}:role\/.+$/, 'Enter a valid AWS IAM role ARN'),
  preset: z.enum(['stacks', 'custom']),
})

export type CreateCloudConnectionValues = z.infer<
  typeof createCloudConnectionSchema
>
