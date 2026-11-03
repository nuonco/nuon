import { z } from 'zod'

const azureUUID = z
  .string()
  .regex(
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    'Enter a UUID in the form 00000000-0000-0000-0000-000000000000'
  )

const accountSchema = z
  .object({
    name: z.string().trim().min(1, 'Connection name is required'),
    platform: z.enum(['aws', 'gcp', 'azure']).optional(),
    target_id: z.string().trim(),
    role_name: z.string().optional(),
    gcp_project_number: z.string().trim().optional(),
    service_account_name: z.string().trim().optional(),
    azure_tenant_id: z.string().trim().optional(),
    principal: z.string().trim().optional(),
    principal_choice: z.string().optional(),
    principal_mode: z.enum(['new', 'existing']).optional(),
  })
  .superRefine((value, ctx) => {
    if (value.principal_mode === 'existing' && !value.principal_choice) {
      ctx.addIssue({
        code: 'custom',
        path: ['principal_choice'],
        message: 'Choose an existing principal',
      })
    }
    const fields =
      value.platform === 'azure'
        ? {
            target_id: azureUUID,
            azure_tenant_id: azureUUID,
            principal: azureUUID,
          }
        : value.platform === 'gcp'
          ? {
              target_id: z
                .string()
                .regex(
                  /^[a-z][a-z0-9-]{4,28}[a-z0-9]$/,
                  'Use a 6–30 character GCP project ID with lowercase letters, numbers, and dashes'
                ),
              gcp_project_number: z
                .string()
                .regex(
                  /^[1-9]\d{5,19}$/,
                  'Enter the numeric GCP project number (6–20 digits, without a leading zero)'
                ),
              service_account_name: z
                .string()
                .regex(
                  /^[a-z][a-z0-9-]{4,28}[a-z0-9]$/,
                  'Use 6–30 lowercase letters, numbers, or dashes; start with a letter and end with a letter or number'
                ),
            }
          : {
              target_id: z
                .string()
                .regex(/^\d{12}$/, 'Enter a 12-digit AWS account ID'),
              role_name: z
                .string()
                .min(1, 'Role name is required')
                .max(64, 'Role name must be at most 64 characters')
                .regex(/^[\w+=,.@-]+$/, 'Use letters, numbers, or _+=,.@-'),
            }
    for (const [field, schema] of Object.entries(fields)) {
      if (field === 'role_name' && value.principal_mode === 'existing') continue
      const result = schema.safeParse(value[field as keyof typeof value] || '')
      if (!result.success) {
        ctx.addIssue({
          code: 'custom',
          path: [field],
          message: result.error.issues[0].message,
        })
      }
    }
  })

const accessSchema = z
  .object({
    platform: z.enum(['aws', 'gcp', 'azure']).optional(),
    preset: z.enum(['stacks', 'ecr', 'gar', 'acr', 'custom']),
    ecr_region: z.string().trim().optional(),
    gar_location: z.string().trim().optional(),
    azure_resource_group: z.string().trim().optional(),
    azure_registry_name: z.string().trim().optional(),
    azure_registry_login_server: z.string().trim().optional(),
  })
  .superRefine((value, ctx) => {
    const cloud = value.platform || 'aws'
    if (cloud === 'azure') {
      if (value.preset !== 'acr') {
        ctx.addIssue({
          code: 'custom',
          path: ['preset'],
          message: 'Azure connections support ACR image repository access only',
        })
        return
      }
      if (
        !/^[a-z0-9_.()-]{1,90}$/i.test(value.azure_resource_group || '') ||
        value.azure_resource_group?.endsWith('.')
      ) {
        ctx.addIssue({
          code: 'custom',
          path: ['azure_resource_group'],
          message:
            'Enter the resource group name, up to 90 letters, numbers, underscores, parentheses, hyphens, or periods; do not end with a period',
        })
      }
      if (!/^[a-z0-9]{5,50}$/i.test(value.azure_registry_name || '')) {
        ctx.addIssue({
          code: 'custom',
          path: ['azure_registry_name'],
          message:
            'Enter the registry resource name: 5–50 letters or numbers, without the login hostname or DNS hash',
        })
      }
      const validLogin =
        /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.azurecr\.io$/.test(
          value.azure_registry_login_server || ''
        )
      if (!validLogin) {
        ctx.addIssue({
          code: 'custom',
          path: ['azure_registry_login_server'],
          message:
            'Use this registry’s bare login server, such as acmeimages.azurecr.io or acmeimages-abc123.azurecr.io',
        })
      }
      return
    }
    if (cloud === 'gcp') {
      if (value.preset !== 'gar') {
        ctx.addIssue({
          code: 'custom',
          path: ['preset'],
          message: 'GCP connections support GAR image repository access only',
        })
        return
      }
      if (!/^[a-z][a-z0-9-]*$/.test(value.gar_location || '')) {
        ctx.addIssue({
          code: 'custom',
          path: ['gar_location'],
          message:
            'Enter the GAR region or multi-region, such as us-central1 or us',
        })
      }
      return
    }
    if (value.preset === 'gar' || value.preset === 'acr') {
      ctx.addIssue({
        code: 'custom',
        path: ['preset'],
        message:
          value.preset === 'acr'
            ? 'ACR access requires an Azure connection'
            : 'GAR access requires a GCP connection',
      })
      return
    }
    if (value.preset !== 'ecr') return
    if (
      !/^[a-z]{2}(?:-[a-z]+)+-\d+$/.test(value.ecr_region || '') ||
      /^(?:us-gov-|cn-)/.test(value.ecr_region || '')
    ) {
      ctx.addIssue({
        code: 'custom',
        path: ['ecr_region'],
        message: 'Choose an AWS region; GovCloud and China are not supported',
      })
    }
  })

export const connectionSchema = accountSchema.and(accessSchema)
export type TConnectionValues = z.infer<typeof connectionSchema>
