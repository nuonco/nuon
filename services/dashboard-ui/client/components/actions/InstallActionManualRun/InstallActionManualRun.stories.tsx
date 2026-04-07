export default {
  title: 'Actions/InstallActionManualRun',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { Text } from '@/components/common/Text'
import { InstallActionManualRunModal } from './InstallActionManualRun'

const noop = () => {}

const mockAction = {
  id: 'action-1',
  name: 'deploy-step',
  configs: [
    {
      steps: [
        { env_vars: { API_URL: 'https://api.example.com', NODE_ENV: 'production' } },
        { env_vars: { DB_HOST: 'db.example.com' } },
      ],
    },
  ],
} as any

export const Default = () => (
  <ModalStory>
    <InstallActionManualRunModal
      action={mockAction}
      actionConfigId="config-1"
      isLoading={false}
      onSubmit={noop}
      roleSelector={<Text variant="subtext">Role selector placeholder</Text>}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <InstallActionManualRunModal
      action={mockAction}
      actionConfigId="config-1"
      isLoading={true}
      onSubmit={noop}
      roleSelector={<Text variant="subtext">Role selector placeholder</Text>}
    />
  </ModalStory>
)

const mockActionNoEnvVars = {
  id: 'action-2',
  name: 'simple-action',
  configs: [{ steps: [{ env_vars: {} }] }],
} as any

export const NoEnvVars = () => (
  <ModalStory>
    <InstallActionManualRunModal
      action={mockActionNoEnvVars}
      actionConfigId="config-2"
      isLoading={false}
      onSubmit={noop}
      roleSelector={<Text variant="subtext">Role selector placeholder</Text>}
    />
  </ModalStory>
)
