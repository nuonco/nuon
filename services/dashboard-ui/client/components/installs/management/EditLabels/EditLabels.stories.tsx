export default {
  title: 'Installs/EditLabels',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EditLabelsModal } from './EditLabels'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'production', team: 'platform', region: 'us-east-1' }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithDefaultLabels = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'production' }}
      defaultLabels={{
        tier: 'gold',
        region: '{{ .nuon.cloud_account.aws.region }}',
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <EditLabelsModal labels={{}} isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'staging' }}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EditLabelsModal
      labels={{ env: 'staging' }}
      isPending={false}
      error={{ error: 'Something went wrong', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
