export default {
  title: 'Installs/EnableConfigSync',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EnableConfigSyncModal } from './EnableConfigSync'

const noop = () => {}

export const EnableDefault = () => (
  <ModalStory>
    <EnableConfigSyncModal isManagedByConfig={false} isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const DisableDefault = () => (
  <ModalStory>
    <EnableConfigSyncModal isManagedByConfig={true} isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <EnableConfigSyncModal isManagedByConfig={false} isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EnableConfigSyncModal isManagedByConfig={false} isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)
