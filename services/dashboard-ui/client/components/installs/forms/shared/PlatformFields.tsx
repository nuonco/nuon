import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import type { IPlatformFields } from './types'

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
}: {
  draftValues?: Record<string, string> | null
  awsAccountConnections?: IPlatformFields['awsAccountConnections']
}) => {
  const options = AWS_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
      : region.text,
  }))

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
          />
        </FieldWrapper>
      ) : null}
    </fieldset>
  )
}

const AzureFields = ({
  draftValues,
}: {
  draftValues?: Record<string, string> | null
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
    </fieldset>
  )
}

export const PlatformFields = ({
  platform,
  draftValues,
  awsAccountConnections,
}: IPlatformFields) => {
  if (platform === 'aws') {
    return (
      <AWSFields
        draftValues={draftValues}
        awsAccountConnections={awsAccountConnections}
      />
    )
  }

  if (platform === 'azure') {
    return <AzureFields draftValues={draftValues} />
  }

  return null
}
