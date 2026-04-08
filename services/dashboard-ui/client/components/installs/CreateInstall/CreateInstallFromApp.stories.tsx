export default {
  title: 'Installs/CreateInstallFromApp',
}

import { CreateInstallFromApp } from './CreateInstallFromApp'

const noop = () => {}
const noopAsync = async () => {}

const mockApp = {
  id: 'app-1',
  name: 'Production App',
  runner_config: { app_runner_type: 'aws' },
} as any

const mockConfig = {
  input: {
    inputs: [],
    input_groups: [],
  },
} as any

export const Default = () => (
  <CreateInstallFromApp
    app={mockApp}
    config={mockConfig}
    isLoading={false}
    error={null}
    isSubmitting={false}
    onSelectApp={noop}
    onClose={noop}
    onSubmit={noopAsync}
  />
)

export const Loading = () => (
  <CreateInstallFromApp
    app={mockApp}
    config={undefined}
    isLoading={true}
    error={null}
    isSubmitting={false}
    onSelectApp={noop}
    onClose={noop}
    onSubmit={noopAsync}
  />
)

export const WithError = () => (
  <CreateInstallFromApp
    app={mockApp}
    config={undefined}
    isLoading={false}
    error={{ error: 'Failed to load config' }}
    isSubmitting={false}
    onSelectApp={noop}
    onClose={noop}
    onSubmit={noopAsync}
  />
)
