import { ModalStory } from '@/components/__stories__/helpers'
import { EditLabelColorsModal } from './EditLabelColors'

export default {
  title: 'Apps/EditLabelColors',
}

const defaultColors = [
  '#2563eb', '#dc2626', '#16a34a', '#9333ea', '#ca8a04', '#0891b2',
  '#e11d48', '#4f46e5', '#059669', '#c026d3', '#d97706', '#0284c7',
  '#7c3aed', '#15803d', '#a21caf', '#b45309', '#6366f1', '#ef4444',
  '#22c55e', '#a855f7', '#eab308', '#06b6d4', '#f43f5e', '#818cf8',
]

export const Empty = () => (
  <ModalStory>
    <EditLabelColorsModal
      labelColors={{}}
      defaultColors={defaultColors}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const WithColors = () => (
  <ModalStory>
    <EditLabelColorsModal
      labelColors={{
        env: '#16a34a',
        region: '#2563eb',
        team: '#9333ea',
      }}
      defaultColors={defaultColors}
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const Saving = () => (
  <ModalStory>
    <EditLabelColorsModal
      labelColors={{ env: '#16a34a' }}
      defaultColors={defaultColors}
      isPending={true}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)
