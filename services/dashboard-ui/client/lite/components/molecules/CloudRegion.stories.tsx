import { ComponentDocs } from '../__stories__/ComponentDocs'
import { CloudRegion } from './CloudRegion'

export default {
  title: 'lite/molecules/CloudRegion',
}

export const Overview = () => (
  <ComponentDocs
    name="CloudRegion"
    tier="molecule"
    summary="A cloud region name paired with the country flag from the shared cloud-region catalog."
    use={[
      'Identify AWS, Azure, and GCP regions in tables, cards, and configuration.',
      'Pass location for Azure and region for AWS or GCP.',
    ]}
    avoid={[
      'Do not duplicate region names or country mappings at the call site.',
      'Do not use it to identify the cloud provider; pair it with CloudPlatform when needed.',
    ]}
    rules={[
      'Region names and flags come from the shared dashboard catalogs.',
      'It renders at caption by default so it matches Time in the same row.',
      'Unknown platforms, missing values, and unrecognized values render Unknown.',
      'The flag is decorative; the readable region name is the accessible content.',
      'Loading is inherited from Text.',
    ]}
    props={[
      {
        name: 'platform',
        type: 'TCloudPlatform',
        description: 'Cloud provider used to select the region catalog.',
      },
      {
        name: 'region',
        type: 'string',
        description: 'AWS or GCP region identifier.',
      },
      {
        name: 'location',
        type: 'string',
        description: 'Azure location identifier.',
      },
      {
        name: 'variant',
        type: 'TTextVariant',
        default: "'caption'",
        description: 'Text size, matched to Time so table rows stay even.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the Text loading state.',
      },
    ]}
  />
)

export const Providers = () => (
  <div className="flex flex-col items-start gap-4 p-8">
    <CloudRegion platform="aws" region="us-west-2" />
    <CloudRegion platform="azure" location="westeurope" />
    <CloudRegion platform="gcp" region="us-central1" />
  </div>
)

export const Unknown = () => (
  <div className="flex flex-col items-start gap-4 p-8">
    <CloudRegion platform="unknown" />
    <CloudRegion platform="aws" region="invalid-region" />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <CloudRegion platform="aws" region="us-west-2" loading loadingWidth={18} />
  </div>
)
