import { z } from 'zod'

const uuid =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

export const createCloudConnectionSchema = z
  .object({
    name: z.string().trim().min(1, 'Name is required'),
    platform: z.enum(['aws', 'azure', 'gcp']),
    targetId: z.string().trim().min(1, 'Target is required'),
    tenantId: z.string(),
    capabilitySet: z.enum(['stacks', 'images', 'both']),
    registry: z.string(),
    repositories: z.string(),
    principal: z.string().trim().min(1, 'Principal is required'),
    identityProvider: z.string(),
  })
  .superRefine((value, context) => {
    if (value.platform === 'aws') {
      if (!/^\d{12}$/.test(value.targetId)) {
        context.addIssue({
          code: 'custom',
          path: ['targetId'],
          message: 'Enter a 12-digit AWS account ID',
        })
      }
      if (!/^arn:aws:iam::\d{12}:role\/.+/.test(value.principal)) {
        context.addIssue({
          code: 'custom',
          path: ['principal'],
          message: 'Enter a valid AWS IAM role ARN',
        })
      }
      return
    }

    if (value.platform === 'gcp') {
      if (!/^[a-z][a-z0-9-]{4,28}[a-z0-9]$/.test(value.targetId)) {
        context.addIssue({
          code: 'custom',
          path: ['targetId'],
          message: 'Enter a valid GCP project ID',
        })
      }
      if (
        !/^[a-z][a-z0-9-]{4,28}[a-z0-9]@[a-z][a-z0-9-]{4,28}[a-z0-9]\.iam\.gserviceaccount\.com$/.test(
          value.principal
        )
      ) {
        context.addIssue({
          code: 'custom',
          path: ['principal'],
          message: 'Enter a valid GCP service account email',
        })
      }
      if (
        !/^(?:https:)?\/\/iam\.googleapis\.com\/projects\/\d+\/locations\/global\/workloadIdentityPools\/[a-z0-9-]+\/providers\/[a-z0-9-]+$|^projects\/\d+\/locations\/global\/workloadIdentityPools\/[a-z0-9-]+\/providers\/[a-z0-9-]+$/.test(
          value.identityProvider.trim()
        )
      ) {
        context.addIssue({
          code: 'custom',
          path: ['identityProvider'],
          message: 'Enter a valid Workload Identity Provider resource name',
        })
      }
      if (value.capabilitySet !== 'images') {
        context.addIssue({
          code: 'custom',
          path: ['capabilitySet'],
          message: 'GCP connections support image pulls only',
        })
      }
      return
    }

    if (!uuid.test(value.tenantId.trim())) {
      context.addIssue({
        code: 'custom',
        path: ['tenantId'],
        message: 'Enter a valid Entra tenant ID',
      })
    }
    if (!uuid.test(value.targetId)) {
      context.addIssue({
        code: 'custom',
        path: ['targetId'],
        message: 'Enter a valid Azure subscription ID',
      })
    }
    if (!uuid.test(value.principal)) {
      context.addIssue({
        code: 'custom',
        path: ['principal'],
        message: 'Enter a valid client ID',
      })
    }
    if (
      value.capabilitySet !== 'stacks' &&
      !/^[a-z0-9]{5,50}$/.test(value.registry.trim())
    ) {
      context.addIssue({
        code: 'custom',
        path: ['registry'],
        message: 'Enter a valid Azure Container Registry name',
      })
    }
  })

export type CreateCloudConnectionValues = z.infer<
  typeof createCloudConnectionSchema
>
