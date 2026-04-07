export default {
  title: 'Installs/EnableAutoApprove',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EnableAutoApproveModal, ConfirmOverrideModal } from './EnableAutoApprove'

const noop = () => {}

export const EnableDefault = () => (
  <ModalStory>
    <EnableAutoApproveModal isPending={false} error={null} isApproveAll={false} isSuccess={false} onSubmit={noop} />
  </ModalStory>
)

export const DisableDefault = () => (
  <ModalStory>
    <EnableAutoApproveModal isPending={false} error={null} isApproveAll={true} isSuccess={false} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <EnableAutoApproveModal isPending={true} error={null} isApproveAll={false} isSuccess={false} onSubmit={noop} />
  </ModalStory>
)

export const Success = () => (
  <ModalStory>
    <EnableAutoApproveModal isPending={false} error={null} isApproveAll={false} isSuccess={true} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EnableAutoApproveModal isPending={false} error={{ error: 'Something went wrong' }} isApproveAll={false} isSuccess={false} onSubmit={noop} />
  </ModalStory>
)

export const ConfirmOverride = () => (
  <ModalStory>
    <ConfirmOverrideModal onConfirm={noop} />
  </ModalStory>
)
