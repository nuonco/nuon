import type { Meta, StoryObj } from '@ladle/react'
import { CloudRegion } from './CloudRegion'

export default {
  title: 'Common/CloudRegion',
} satisfies Meta

export const AWS: StoryObj = {
  render: () => <CloudRegion platform="aws" region="us-east-1" />,
}

export const Azure: StoryObj = {
  render: () => <CloudRegion platform="azure" location="eastus" />,
}

export const GCP: StoryObj = {
  render: () => <CloudRegion platform="gcp" region="us-central1" />,
}

export const Unknown: StoryObj = {
  render: () => <CloudRegion platform="aws" region="invalid-region" />,
}
