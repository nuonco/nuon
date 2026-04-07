export default {
  title: 'Installs/LoadAppConfigs',
}

import { LoadAppConfigs } from './LoadAppConfigs'
import { Text } from '@/components/common/Text'

const noop = () => {}

const mockApp = {
  id: 'app-1',
  name: 'Production App',
  runner_config: { app_runner_type: 'aws' },
} as any

const mockConfigs = [{ id: 'config-1' }] as any[]

export const WithConfigs = () => (
  <LoadAppConfigs
    app={mockApp}
    configs={mockConfigs}
    isLoading={false}
    error={null}
    onSelectApp={noop}
  >
    <Text>Config loaded content goes here</Text>
  </LoadAppConfigs>
)

export const Loading = () => (
  <LoadAppConfigs
    app={mockApp}
    configs={undefined}
    isLoading={true}
    error={null}
    onSelectApp={noop}
  />
)

export const WithError = () => (
  <LoadAppConfigs
    app={mockApp}
    configs={undefined}
    isLoading={false}
    error={{ error: 'Unable to load configs' }}
    onSelectApp={noop}
  />
)

export const Empty = () => (
  <LoadAppConfigs
    app={mockApp}
    configs={[]}
    isLoading={false}
    error={null}
    onSelectApp={noop}
  />
)
