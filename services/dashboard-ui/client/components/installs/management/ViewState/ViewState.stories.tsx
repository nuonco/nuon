export default {
  title: 'Installs/ViewState',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ViewStateModal } from './ViewState'

const mockState = {
  runner: { status: 'active', version: '1.2.3' },
  sandbox: { status: 'provisioned' },
  components: [{ name: 'web', status: 'deployed' }],
}

export const Default = () => (
  <ModalStory>
    <ViewStateModal state={mockState} error={null} isLoading={false} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ViewStateModal state={null} error={null} isLoading={true} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ViewStateModal state={null} error={{ error: 'Unable to load state' }} isLoading={false} />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <ViewStateModal state={null} error={null} isLoading={false} />
  </ModalStory>
)
