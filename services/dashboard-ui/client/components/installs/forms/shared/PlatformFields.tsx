import { useState } from 'react'

import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import type { IPlatformFields } from './types'

export const AWS_ACCOUNT_ID_PATTERN = '[0-9]{12}'

const FieldWrapper = ({
  children,
  labelText,
  helpText,
  required = true,
}: {
  children: React.ReactElement
  labelText: string
  helpText?: string
  required?: boolean
}) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">
          {labelText}{' '}
          {required ? (
            <Text className="ml-1" variant="subtext" theme="error">
              {'*'}
            </Text>
          ) : null}
        </Text>
        {helpText ? (
          <Text variant="subtext" className="max-w-72">
            {helpText}
          </Text>
        ) : null}
      </span>
      {children}
    </label>
  )
}

const AWSFields = ({
  draftValues,
  awsAccountConnections,
  requireTargetAccount,
}: {
  draftValues?: Record<string, string> | null
  awsAccountConnections?: IPlatformFields['awsAccountConnections']
  requireTargetAccount?: boolean
}) => {
  const options = AWS_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
      : region.text,
  }))

  const [connectionId, setConnectionId] = useState(
    draftValues?.aws_connection_id || ''
  )
  const connectionAccountId = awsAccountConnections?.find(
    (connection) => connection.id === connectionId
  )?.account_id

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-3 pr-6">
        Set AWS settings{' '}
        <Text className="ml-1" variant="subtext" theme="error">
          (required)
        </Text>
      </legend>

      <FieldWrapper labelText="Select AWS region">
        <Select
          name="region"
          options={options}
          placeholder="Choose AWS region"
          required
          searchable
          defaultValue={draftValues?.region || ''}
        />
      </FieldWrapper>

      {awsAccountConnections ? (
        <FieldWrapper
          labelText="AWS connection (optional)"
          helpText="Select an AWS connection for Nuon to apply the install stack. Leave as None if the customer will apply it."
          required={false}
        >
          <Select
            name="aws_connection_id"
            options={[
              {
                value: '',
                label: 'None — customer will apply the stack',
              },
              ...awsAccountConnections.map((connection) => ({
                value: connection.id,
                label: `${connection.name} · ${connection.account_id} · ${connection.verification_status === 'verified' ? 'Verified' : connection.verification_status}`,
                disabled: connection.verification_status !== 'verified',
              })),
            ]}
            defaultValue={draftValues?.aws_connection_id || ''}
            onChange={(value) => setConnectionId(value)}
          />
        </FieldWrapper>
      ) : null}

      <FieldWrapper
        labelText="AWS account ID"
        helpText={
          connectionAccountId
            ? 'Taken from the selected AWS connection.'
            : 'The AWS account this install is deployed into. Cannot be changed later.'
        }
        required={!!requireTargetAccount}
      >
        <Input
          name="aws_account_id"
          placeholder="123456789012"
          pattern={AWS_ACCOUNT_ID_PATTERN}
          title="Enter a 12-digit AWS account ID"
          inputMode="numeric"
          required={!!requireTargetAccount}
          readOnly={!!connectionAccountId}
          key={connectionAccountId || 'manual'}
          defaultValue={connectionAccountId || draftValues?.aws_account_id || ''}
        />
      </FieldWrapper>
    </fieldset>
  )
}

const AzureFields = ({
  draftValues,
  requireTargetAccount,
}: {
  draftValues?: Record<string, string> | null
  requireTargetAccount?: boolean
}) => {
  const options = AZURE_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text}`
      : region.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-3 pr-6">
        Set Azure configuration{' '}
        <Text className="ml-1" variant="subtext" theme="error">
          (required)
        </Text>
      </legend>

      <FieldWrapper labelText="Select Azure location">
        <Select
          name="location"
          options={options}
          placeholder="Choose Azure location"
          required
          searchable
          defaultValue={draftValues?.location || ''}
        />
      </FieldWrapper>

      <FieldWrapper
        labelText="Azure subscription ID"
        helpText="The Azure subscription this install is deployed into. Cannot be changed later."
        required={!!requireTargetAccount}
      >
        <Input
          name="azure_subscription_id"
          placeholder="00000000-0000-0000-0000-000000000000"
          required={!!requireTargetAccount}
          defaultValue={draftValues?.azure_subscription_id || ''}
        />
      </FieldWrapper>
    </fieldset>
  )
}

const GCPFields = ({
  draftValues,
  requireTargetAccount,
}: {
  draftValues?: Record<string, string> | null
  requireTargetAccount?: boolean
}) => {
  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-3 pr-6">
        Set GCP configuration{' '}
        <Text className="ml-1" variant="subtext" theme="error">
          (required)
        </Text>
      </legend>

      <FieldWrapper
        labelText="GCP project ID"
        helpText="The GCP project this install is deployed into. Cannot be changed later."
        required={!!requireTargetAccount}
      >
        <Input
          name="gcp_project_id"
          placeholder="my-gcp-project"
          required={!!requireTargetAccount}
          defaultValue={draftValues?.gcp_project_id || ''}
        />
      </FieldWrapper>
    </fieldset>
  )
}

export const PlatformFields = ({
  platform,
  draftValues,
  awsAccountConnections,
  requireTargetAccount,
}: IPlatformFields) => {
  if (platform === 'aws') {
    return (
      <AWSFields
        draftValues={draftValues}
        awsAccountConnections={awsAccountConnections}
        requireTargetAccount={requireTargetAccount}
      />
    )
  }

  if (platform === 'azure') {
    return (
      <AzureFields
        draftValues={draftValues}
        requireTargetAccount={requireTargetAccount}
      />
    )
  }

  if (platform === 'gcp') {
    return (
      <GCPFields
        draftValues={draftValues}
        requireTargetAccount={requireTargetAccount}
      />
    )
  }

  return null
}
