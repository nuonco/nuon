export default {
  title: 'Workflows/StepDetails/AwaitAzureDetails',
}

import { AwaitAzureDetails, AwaitAzureDetailsSkeleton } from './AwaitAzureDetails'

const mockStack = {
  versions: [
    {
      template_url: 'https://storage.azure.com/template.json',
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

const mockSecrets = [
  {
    name: 'optional_token',
    display_name: 'Optional token',
    description: 'Only needed when SSO is enabled.',
  },
  {
    name: 'license_key',
    display_name: 'License key',
    description: 'The license key for your app.',
    required: true,
  },
  {
    name: 'smtp_password',
    display_name: 'SMTP password',
    default: 'change-me',
  },
  {
    name: 'session_signing_key',
    display_name: 'Session signing key',
    auto_generate: true,
  },
  {
    name: 'tls_cert',
    display_name: 'TLS certificate',
    required: true,
    format: 'base64',
  },
] as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
    />
  </div>
)

export const WithSecrets = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      azureLocation="eastus"
      secrets={mockSecrets}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitAzureDetailsSkeleton />
  </div>
)
