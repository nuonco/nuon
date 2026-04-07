export default {
  title: 'Installs/UpdateInstallForm',
}

import { UpdateInstallForm } from './UpdateInstallForm'

const mockInstall = {
  id: 'install-1',
  name: 'my-install',
} as any

const mockInputConfig = {
  id: 'config-1',
  input_values: [
    { name: 'api_key', type: 'string', required: true },
    { name: 'region', type: 'string', required: false },
  ],
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <UpdateInstallForm
      install={mockInstall}
      platform="aws"
      inputConfig={mockInputConfig}
      onCancel={() => {}}
    />
  </div>
)

export const WithoutInputConfig = () => (
  <div className="max-w-2xl p-4">
    <UpdateInstallForm
      install={mockInstall}
      platform="aws"
      onCancel={() => {}}
    />
  </div>
)
