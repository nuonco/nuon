import { z } from 'zod'
import type { TAppInput, TAppInputConfig } from '@/types'

export type InstallFormMode = 'create' | 'edit'
export type InstallPlatform = 'aws' | 'azure' | 'gcp'
export type InstallInputFieldName = `inputs.${string}`

export const normalizeInstallPlatform = (
  runnerType?: string
): InstallPlatform | undefined => {
  if (!runnerType) return undefined
  if (runnerType.startsWith('aws')) return 'aws'
  if (runnerType.startsWith('azure')) return 'azure'
  if (runnerType.startsWith('gcp')) return 'gcp'
  return undefined
}

export interface InstallFormValues {
  name: string
  region: string
  aws_connection_id: string
  aws_account_id: string
  location: string
  azure_subscription_id: string
  gcp_project_id: string
  autoApprove: boolean
  vpcTemplateUrl: string
  runnerTemplateUrl: string
  labels: { key: string; value: string }[]
  role: string
  deployDependents: boolean
  inputs: Record<string, string | boolean>
}

export const AWS_ACCOUNT_ID_REGEX = /^[0-9]{12}$/

export const isBooleanInput = (input: TAppInput): boolean =>
  input?.type === 'bool' ||
  input?.default === 'true' ||
  input?.default === 'false'

export const getEditableInputs = (
  inputConfig?: TAppInputConfig
): TAppInput[] => {
  if (!inputConfig?.input_groups) return []
  return inputConfig.input_groups.flatMap((group) =>
    (group?.app_inputs ?? []).filter((input) => input?.source !== 'customer')
  )
}

const inputFieldSchema = (input: TAppInput): z.ZodTypeAny => {
  if (isBooleanInput(input)) {
    return z.boolean()
  }
  if (input?.required) {
    const label = input?.display_name || input?.name || 'This field'
    return z
      .string()
      .trim()
      .min(1, `${label} is required`)
  }
  return z.string()
}

export interface BuildInstallSchemaParams {
  mode: InstallFormMode
  platform?: InstallPlatform
  inputConfig?: TAppInputConfig
  requireTargetAccount?: boolean
  showNameField?: boolean
}

export const buildInstallSchema = ({
  mode,
  platform,
  inputConfig,
  requireTargetAccount,
  showNameField,
}: BuildInstallSchemaParams) => {
  const inputsShape: Record<string, z.ZodTypeAny> = {}
  for (const input of getEditableInputs(inputConfig)) {
    if (!input?.name) continue
    inputsShape[input.name] = inputFieldSchema(input)
  }

  const shape: Record<string, z.ZodTypeAny> = {
    inputs: z.object(inputsShape),
  }

  if (mode === 'create' || showNameField) {
    shape.name = z.string().trim().min(1, 'Install name is required')
  } else {
    shape.name = z.string()
  }

  if (mode === 'create') {
    if (platform === 'aws') {
      shape.region = z.string().min(1, 'Select an AWS region')
      shape.aws_account_id = requireTargetAccount
        ? z
            .string()
            .regex(AWS_ACCOUNT_ID_REGEX, 'Enter a 12-digit AWS account ID')
        : z.union([
            z.literal(''),
            z
              .string()
              .regex(AWS_ACCOUNT_ID_REGEX, 'Enter a 12-digit AWS account ID'),
          ])
    } else if (platform === 'azure') {
      shape.location = z.string().min(1, 'Select an Azure location')
      if (requireTargetAccount) {
        shape.azure_subscription_id = z
          .string()
          .trim()
          .min(1, 'Azure subscription ID is required')
      }
    } else if (platform === 'gcp') {
      if (requireTargetAccount) {
        shape.gcp_project_id = z
          .string()
          .trim()
          .min(1, 'GCP project ID is required')
      }
    }
  }

  return z.object(shape)
}
