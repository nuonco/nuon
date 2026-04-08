export default {
  title: 'Installs/ViewCurrentInputs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ViewCurrentInputsModal } from './ViewCurrentInputs'

const mockInputGroups = [
  {
    id: 'group-1',
    display_name: 'Database settings',
    description: 'Configure database connection',
    app_inputs: [
      { name: 'db_host', display_name: 'Database host', description: 'Hostname', required: true, sensitive: false, source: 'vendor', default: 'localhost' },
      { name: 'db_password', display_name: 'Database password', description: 'Password', required: true, sensitive: true, source: 'customer' },
    ],
  },
]

const mockRedacted = {
  db_host: 'prod-db.example.com',
  db_password: '****',
}

export const WithGroups = () => (
  <ModalStory>
    <ViewCurrentInputsModal isLoading={false} redactedValues={mockRedacted} inputGroups={mockInputGroups} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ViewCurrentInputsModal isLoading={true} redactedValues={{}} inputGroups={[]} />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <ViewCurrentInputsModal isLoading={false} redactedValues={{}} inputGroups={[]} />
  </ModalStory>
)

export const FlatInputs = () => (
  <ModalStory>
    <ViewCurrentInputsModal isLoading={false} redactedValues={{ key1: 'value1', key2: 'value2' }} inputGroups={[]} />
  </ModalStory>
)
