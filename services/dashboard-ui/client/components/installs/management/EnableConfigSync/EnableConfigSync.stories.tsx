export default {
  title: 'Installs/EnableConfigSync',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EnableConfigSyncModal, DisableConfigSyncModal } from './EnableConfigSync'

const noop = () => {}

export const Enable = () => (
  <ModalStory>
    <EnableConfigSyncModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const EnableLoading = () => (
  <ModalStory>
    <EnableConfigSyncModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const EnableError = () => (
  <ModalStory>
    <EnableConfigSyncModal isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)

export const Disable = () => (
  <ModalStory>
    <DisableConfigSyncModal installName="prod-acme" isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const DisableError = () => (
  <ModalStory>
    <DisableConfigSyncModal installName="prod-acme" isPending={false} error={{ error: 'Something went wrong' }} onSubmit={noop} />
  </ModalStory>
)
