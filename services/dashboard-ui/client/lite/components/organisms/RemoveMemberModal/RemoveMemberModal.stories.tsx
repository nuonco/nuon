import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { RemoveMemberModal } from './RemoveMemberModal'

export default { title: 'lite/organisms/RemoveMemberModal' }

const email = 'member@example.com'

export const Overview = () => (
  <ComponentDocs
    name="RemoveMemberModal"
    tier="organism"
    summary="A type-to-confirm decision that removes one active member from the org."
    use={['Open it from an active member row before revoking org access.']}
    avoid={[
      'Do not use it for pending invites.',
      'Do not allow submission until the full email matches.',
    ]}
    rules={[
      'The consequence, warning, and required confirmation value remain visible together.',
      'Mutation errors stay in the modal.',
    ]}
    props={[
      { name: 'email', type: 'string', description: 'Member being removed.' },
      {
        name: 'confirmation',
        type: 'string',
        description: 'Current type-to-confirm value.',
      },
      {
        name: 'onConfirmationChange',
        type: '(value: string) => void',
        description: 'Updates the type-to-confirm value.',
      },
      {
        name: 'onSubmit',
        type: '() => void',
        description: 'Removes the member after confirmation.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Locks the action while removal is in progress.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Displays a mutation failure inside the modal.',
      },
    ]}
  />
)

export const Blocked = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <RemoveMemberModal
          email={email}
          confirmation=""
          onConfirmationChange={() => {}}
          onSubmit={() => {}}
        />
      )
    }
  />
)

export const Unblocked = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <RemoveMemberModal
          email={email}
          confirmation={email}
          onConfirmationChange={() => {}}
          onSubmit={() => {}}
        />
      )
    }
  />
)
