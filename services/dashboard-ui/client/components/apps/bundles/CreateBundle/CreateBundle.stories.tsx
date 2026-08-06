import { ModalStory } from '@/components/__stories__/helpers'
import { CreateBundleModal } from './CreateBundle'

export default {
  title: 'Apps/Bundles/CreateBundle',
}

export const Default = () => (
  <ModalStory>
    <CreateBundleModal
      appName="my-app"
      appConfigId="app608mmtt1p456k8f65znhl43"
      isPending={false}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateBundleModal
      appName="my-app"
      appConfigId="app608mmtt1p456k8f65znhl43"
      isPending={true}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const NoConfig = () => (
  <ModalStory>
    <CreateBundleModal appName="my-app" isPending={false} onSubmit={() => {}} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateBundleModal
      appName="my-app"
      appConfigId="app608mmtt1p456k8f65znhl43"
      isPending={false}
      error={{ error: 'airgap bundle storage is not configured' }}
      onSubmit={() => {}}
    />
  </ModalStory>
)
