import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { RevokeInviteModal } from './RevokeInviteModal'

export default { title: 'lite/organisms/RevokeInviteModal' }

export const Overview = () => (
  <ComponentDocs
    name="RevokeInviteModal"
    tier="organism"
    summary="A destructive confirmation that prevents one pending invite from being accepted."
    use={['Open it from a pending invite row when the caller is an org admin.']}
    avoid={[
      'Do not offer it to non-admins.',
      'Do not use it for active members.',
    ]}
    rules={[
      'The consequence and invitee email are visible before revocation.',
      'Mutation errors stay in the modal.',
    ]}
    props={[
      { name: 'email', type: 'string', description: 'Invitee being revoked.' },
      {
        name: 'onSubmit',
        type: '() => void',
        description: 'Revokes the invitation.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Locks the action while revocation is in progress.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Displays a mutation failure inside the modal.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <RevokeInviteModal email="invitee@example.com" onSubmit={() => {}} />
      )
    }
  />
)
