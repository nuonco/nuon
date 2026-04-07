export default {
  title: 'Installs/GenerateInstallConfig',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { GenerateInstallConfigModal } from './GenerateInstallConfig'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <GenerateInstallConfigModal content={'[install]\nname = "prod"'} error={null} isLoading={false} onDownload={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <GenerateInstallConfigModal content={undefined} error={null} isLoading={true} onDownload={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <GenerateInstallConfigModal content={undefined} error={{ error: 'Unable to generate config' }} isLoading={false} onDownload={noop} />
  </ModalStory>
)
