import { Select } from '@/components/common/form/Select'
import { Text } from '@/components/common/Text'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import type { IPlatformFields } from './types'

const FieldWrapper = ({ children, labelText, helpText }: {
  children: React.ReactElement
  labelText: string
  helpText?: string
}) => {
  return (
    <label className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start">
      <span className="flex flex-col gap-0">
        <Text variant="body" weight="strong">{labelText}</Text>
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

const AWSFields = () => {
  const options = AWS_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text} [${region.value}]`
      : region.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-6 pr-6">
        Set AWS settings
      </legend>
      
      <FieldWrapper labelText="Select AWS region *">
        <Select 
          name="region" 
          options={options} 
          placeholder="Choose AWS region"
          required 
        />
      </FieldWrapper>
    </fieldset>
  )
}

const AzureFields = () => {
  const options = AZURE_REGIONS.map((region) => ({
    value: region.value,
    label: region?.iconVariant
      ? `${getFlagEmoji(region.iconVariant.substring(5))} ${region.text}`
      : region.text,
  }))

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="text-lg font-semibold mb-6 pr-6">
        Set Azure configuration
      </legend>
      
      <FieldWrapper labelText="Select Azure location *">
        <Select 
          name="location" 
          options={options} 
          placeholder="Choose Azure location"
          required 
        />
      </FieldWrapper>
    </fieldset>
  )
}

export const PlatformFields = ({ platform }: IPlatformFields) => {
  if (platform === 'aws') {
    return <AWSFields />
  }
  
  if (platform === 'azure') {
    return <AzureFields />
  }
  
  return null
}