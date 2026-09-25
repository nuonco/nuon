import { z } from 'zod'

export const createCloudConnectionSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  platform: z.literal('aws'),
  targetId: z
    .string()
    .trim()
    .regex(/^\d{12}$/, 'Enter a 12-digit AWS account ID'),
  capabilitySet: z.enum(['stacks', 'images', 'both']),
  repositories: z.string(),
  principal: z
    .string()
    .trim()
    .regex(/^arn:aws:iam::\d{12}:role\/.+/, 'Enter a valid AWS IAM role ARN'),
})

export type CreateCloudConnectionValues = z.infer<
  typeof createCloudConnectionSchema
>
