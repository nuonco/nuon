import { ModalStory } from '@/components/__stories__/helpers'
import { ToggleComponentModal } from './ToggleComponent'

export default {
  title: 'Install Components/ToggleComponent',
}

const mockComponent = {
  id: 'cmp123',
  name: 'redis',
  type: 'helm_chart' as const,
}

export const Enable = () => (
  <ModalStory>
    <ToggleComponentModal
      component={mockComponent as any}
      enabling={true}
      isPending={false}
      onSubmit={() => {}}
      onClose={() => {}}
      roleSelector={() => null}
    />
  </ModalStory>
)

export const Disable = () => (
  <ModalStory>
    <ToggleComponentModal
      component={mockComponent as any}
      enabling={false}
      isPending={false}
      onSubmit={() => {}}
      onClose={() => {}}
      roleSelector={() => null}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <ToggleComponentModal
      component={mockComponent as any}
      enabling={true}
      isPending={true}
      onSubmit={() => {}}
      onClose={() => {}}
      roleSelector={() => null}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ToggleComponentModal
      component={mockComponent as any}
      enabling={true}
      isPending={false}
      error={{ error: 'Component is not toggleable' }}
      onSubmit={() => {}}
      onClose={() => {}}
      roleSelector={() => null}
    />
  </ModalStory>
)
