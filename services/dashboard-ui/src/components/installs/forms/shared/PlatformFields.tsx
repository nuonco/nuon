import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { AWS_REGIONS, AZURE_REGIONS, GCP_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import type { IPlatformFields } from './types'

const FieldWrapper = ({
  children,
  labelText,
  helpText,
}: {
  children: React.ReactElement
  labelText: string
  helpText?: string
}) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">
          {labelText}{' '}
          <Text className="ml-1" variant="subtext" theme="error">
            {'*'}
          </Text>
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
}: {
  draftValues?: Record<string, string> | null
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
          defaultValue={draftValues?.region || ''}
        />
      </FieldWrapper>
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
          defaultValue={draftValues?.location || ''}
        />
      </FieldWrapper>
    </fieldset>
  )
}

const GCPFields = ({
  draftValues,
}: {
  draftValues?: Record<string, string> | null
}) => {
  const options = GCP_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
      : region.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-3 pr-6">
        Set GCP configuration{' '}
        <Text className="ml-1" variant="subtext" theme="error">
          (required)
        </Text>
      </legend>

      <FieldWrapper labelText="GCP Project ID">
        <input
          name="project_id"
          type="text"
          placeholder="my-gcp-project"
          required
          defaultValue={draftValues?.project_id || ''}
          className="border rounded px-3 py-2 text-sm"
        />
      </FieldWrapper>

      <FieldWrapper labelText="Select GCP region">
        <Select
          name="gcp_region"
          options={options}
          placeholder="Choose GCP region"
          required
          defaultValue={draftValues?.gcp_region || ''}
        />
      </FieldWrapper>
    </fieldset>
  )
}

export const PlatformFields = ({
  platform,
  draftValues,
}: IPlatformFields) => {
  if (platform === 'aws') {
    return <AWSFields draftValues={draftValues} />
  }

  if (platform === 'azure') {
    return <AzureFields draftValues={draftValues} />
  }

  if (platform === 'gcp') {
    return <GCPFields draftValues={draftValues} />
  }

  return null
}
