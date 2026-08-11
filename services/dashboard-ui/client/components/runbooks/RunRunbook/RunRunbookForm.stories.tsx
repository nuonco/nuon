export default {
  title: 'Runbooks/RunRunbook',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { Text } from '@/components/common/Text'
import { RunRunbookForm } from './RunRunbookForm'

const noop = () => {}

const roleSelector = <Text variant="subtext">Role selector placeholder</Text>

const withInputs = {
  id: 'irb-1',
  runbook_id: 'rb-1',
  runbook: {
    name: 'rotate-secrets',
    configs: [
      {
        inputs: [
          {
            id: 'in-1',
            name: 'target',
            display_name: 'Target',
            required: true,
            type: 'string',
            idx: 0,
          },
          {
            id: 'in-2',
            name: 'dry_run',
            display_name: 'Dry run',
            type: 'bool',
            default: 'true',
            idx: 1,
          },
        ],
        steps: [
          { id: 's1', name: 'Plan', type: 'terraform', idx: 0 },
          { id: 's2', name: 'Apply', type: 'terraform', idx: 1 },
        ],
      },
    ],
  },
} as any

const noInputs = {
  id: 'irb-2',
  runbook_id: 'rb-2',
  runbook: {
    name: 'restart-workers',
    configs: [{ inputs: [], steps: [{ id: 's1', name: 'Restart', type: 'helm', idx: 0 }] }],
  },
} as any

export const WithInputs = () => (
  <ModalStory>
    <RunRunbookForm
      installRunbook={withInputs}
      isPending={false}
      error={null}
      onSubmit={noop}
      roleSelector={roleSelector}
    />
  </ModalStory>
)

export const NoInputs = () => (
  <ModalStory>
    <RunRunbookForm
      installRunbook={noInputs}
      isPending={false}
      error={null}
      onSubmit={noop}
      roleSelector={roleSelector}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RunRunbookForm
      installRunbook={withInputs}
      isPending={false}
      error={{ error: 'Unable to run rotate-secrets', description: '', user_error: true }}
      onSubmit={noop}
      roleSelector={roleSelector}
    />
  </ModalStory>
)
