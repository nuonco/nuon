import type { TCloudPlatform } from '@/types'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Text } from '../atoms/Text'
import { CloudPlatform } from './CloudPlatform'

export default {
  title: 'lite/molecules/CloudPlatform',
}

const PLATFORMS: TCloudPlatform[] = ['aws', 'azure', 'gcp', 'unknown']

export const Overview = () => (
  <ComponentDocs
    name="CloudPlatform"
    tier="molecule"
    summary="A cloud provider mark and label with color-first local artwork."
    use={[
      'Identify the cloud provider in tables, cards, and configuration.',
      'Use icon display only in dense surfaces with clear surrounding context.',
    ]}
    avoid={[
      'Do not import cloud icons directly.',
      'Do not use mono unless the surrounding surface needs one text color.',
    ]}
    rules={[
      'Color is the default.',
      'Unknown platforms use the Phosphor question mark.',
      'Icon-only display includes a keyboard-accessible tooltip.',
    ]}
    props={[
      {
        name: 'platform',
        type: 'TCloudPlatform',
        description: 'Cloud provider to display.',
      },
      {
        name: 'display',
        type: "'abbr' | 'name' | 'icon'",
        default: "'abbr'",
        description: 'Label form or icon-only display.',
      },
      {
        name: 'tone',
        type: "'color' | 'mono'",
        default: "'color'",
        description: 'Full-color artwork or currentColor.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="flex flex-wrap items-center gap-6 p-8">
    {PLATFORMS.map((platform) => (
      <CloudPlatform key={platform} platform={platform} />
    ))}
  </div>
)

export const Names = () => (
  <div className="flex flex-wrap items-center gap-6 p-8">
    {PLATFORMS.map((platform) => (
      <CloudPlatform key={platform} platform={platform} display="name" />
    ))}
  </div>
)

export const Icons = () => (
  <div className="flex items-center gap-6 p-8">
    {PLATFORMS.map((platform) => (
      <CloudPlatform
        key={platform}
        platform={platform}
        display="icon"
        iconSize={24}
      />
    ))}
  </div>
)

export const Mono = () => (
  <div className="flex flex-wrap items-center gap-6 p-8 text-secondary">
    {PLATFORMS.map((platform) => (
      <CloudPlatform key={platform} platform={platform} tone="mono" />
    ))}
  </div>
)

export const InText = () => (
  <div className="flex max-w-xl flex-col gap-3 p-8">
    <Text>
      Deploying to <CloudPlatform platform="aws" variant="caption" /> in
      us-west-2.
    </Text>
    <CloudPlatform
      platform="azure"
      display="name"
      variant="heading"
      weight="semibold"
      iconSize={24}
    />
  </div>
)
