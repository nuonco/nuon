export default {
  title: 'Installs/CreateInstallForm',
}

import { CreateInstallForm } from './CreateInstallForm'

const mockInputConfig = {
  id: 'config-1',
  input_values: [
    { name: 'api_key', type: 'string', required: true },
    { name: 'region', type: 'string', required: false },
  ],
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <CreateInstallForm
      appId="app-1"
      platform="aws"
      inputConfig={mockInputConfig}
      onCancel={() => {}}
    />
  </div>
)

export const WithoutInputConfig = () => (
  <div className="max-w-2xl p-4">
    <CreateInstallForm
      appId="app-1"
      platform="aws"
      onCancel={() => {}}
    />
  </div>
)
