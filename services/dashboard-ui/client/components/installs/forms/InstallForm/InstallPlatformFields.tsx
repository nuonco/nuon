import { useEffect } from 'react'
import { useStore } from '@tanstack/react-form'
import { FormInput } from '@/components/common/form/FormInput'
import { FormSelect } from '@/components/common/form/FormSelect'
import { Text } from '@/components/common/Text'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import type { TAWSAccountConnection } from '@/types'
import { FieldRow } from './FieldRow'
import type { InstallFormApi } from './useInstallForm'
import type { InstallPlatform } from './schema'

interface IInstallPlatformFields {
  form: InstallFormApi
  platform: InstallPlatform
  awsAccountConnections?: TAWSAccountConnection[]
  requireTargetAccount?: boolean
  disabled?: boolean
}

const PlatformLegend = ({ children }: { children: string }) => (
  <legend className="text-lg font-semibold mb-3 pr-6">
    {children}{' '}
    <Text className="ml-1" variant="subtext" theme="error">
      (required)
    </Text>
  </legend>
)

const AwsFields = ({
  form,
  awsAccountConnections,
  requireTargetAccount,
  disabled,
}: Omit<IInstallPlatformFields, 'platform'>) => {
  const regionOptions = AWS_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
      : region.text,
  }))

  const connectionId = useStore(form.store, (s) => s.values.aws_connection_id)
  const connectionAccountId = awsAccountConnections?.find(
    (connection) => connection.id === connectionId
  )?.account_id

  useEffect(() => {
    if (connectionAccountId) {
      form.setFieldValue('aws_account_id', connectionAccountId)
    }
  }, [connectionAccountId, form])

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <PlatformLegend>Set AWS settings</PlatformLegend>

      <FieldRow labelText="Select AWS region" required>
        <form.Field name="region">
          {(field) => (
            <FormSelect
              field={field}
              options={regionOptions}
              placeholder="Choose AWS region"
              searchable
              disabled={disabled}
            />
          )}
        </form.Field>
      </FieldRow>

      {awsAccountConnections ? (
        <FieldRow
          labelText="AWS connection"
          optional
          helpText="Select an AWS connection for Nuon to apply the install stack. Leave as None if the customer will apply it."
        >
          <form.Field name="aws_connection_id">
            {(field) => (
              <FormSelect
                field={field}
                options={[
                  { value: '', label: 'None — customer will apply the stack' },
                  ...awsAccountConnections.map((connection) => ({
                    value: connection.id,
                    label: `${connection.name} · ${connection.account_id} · ${connection.verification_status === 'verified' ? 'Verified' : connection.verification_status}`,
                    disabled: connection.verification_status !== 'verified',
                  })),
                ]}
                disabled={disabled}
              />
            )}
          </form.Field>
        </FieldRow>
      ) : null}

      <FieldRow
        labelText="AWS account ID"
        required={!!requireTargetAccount}
        helpText={
          connectionAccountId
            ? 'Taken from the selected AWS connection.'
            : 'The AWS account this install is deployed into. Cannot be changed later.'
        }
      >
        <form.Field name="aws_account_id">
          {(field) => (
            <FormInput
              field={field}
              placeholder="123456789012"
              inputMode="numeric"
              readOnly={!!connectionAccountId}
              disabled={disabled}
            />
          )}
        </form.Field>
      </FieldRow>
    </fieldset>
  )
}

const AzureFields = ({
  form,
  requireTargetAccount,
  disabled,
}: Omit<IInstallPlatformFields, 'platform' | 'awsAccountConnections'>) => {
  const locationOptions = AZURE_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text}`
      : region.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <PlatformLegend>Set Azure configuration</PlatformLegend>

      <FieldRow labelText="Select Azure location" required>
        <form.Field name="location">
          {(field) => (
            <FormSelect
              field={field}
              options={locationOptions}
              placeholder="Choose Azure location"
              searchable
              disabled={disabled}
            />
          )}
        </form.Field>
      </FieldRow>

      <FieldRow
        labelText="Azure subscription ID"
        required={!!requireTargetAccount}
        helpText="The Azure subscription this install is deployed into. Cannot be changed later."
      >
        <form.Field name="azure_subscription_id">
          {(field) => (
            <FormInput
              field={field}
              placeholder="00000000-0000-0000-0000-000000000000"
              disabled={disabled}
            />
          )}
        </form.Field>
      </FieldRow>
    </fieldset>
  )
}

const GcpFields = ({
  form,
  requireTargetAccount,
  disabled,
}: Omit<IInstallPlatformFields, 'platform' | 'awsAccountConnections'>) => (
  <fieldset className="flex flex-col gap-6 border-t pt-6">
    <PlatformLegend>Set GCP configuration</PlatformLegend>

    <FieldRow
      labelText="GCP project ID"
      required={!!requireTargetAccount}
      helpText="The GCP project this install is deployed into. Cannot be changed later."
    >
      <form.Field name="gcp_project_id">
        {(field) => (
          <FormInput field={field} placeholder="my-gcp-project" disabled={disabled} />
        )}
      </form.Field>
    </FieldRow>
  </fieldset>
)

export const InstallPlatformFields = ({
  form,
  platform,
  awsAccountConnections,
  requireTargetAccount,
  disabled,
}: IInstallPlatformFields) => {
  if (platform === 'aws') {
    return (
      <AwsFields
        form={form}
        awsAccountConnections={awsAccountConnections}
        requireTargetAccount={requireTargetAccount}
        disabled={disabled}
      />
    )
  }
  if (platform === 'azure') {
    return (
      <AzureFields
        form={form}
        requireTargetAccount={requireTargetAccount}
        disabled={disabled}
      />
    )
  }
  if (platform === 'gcp') {
    return (
      <GcpFields
        form={form}
        requireTargetAccount={requireTargetAccount}
        disabled={disabled}
      />
    )
  }
  return null
}
