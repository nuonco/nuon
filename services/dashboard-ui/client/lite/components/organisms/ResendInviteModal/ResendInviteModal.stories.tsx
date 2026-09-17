import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { ResendInviteModal } from './ResendInviteModal'

export default { title: 'lite/organisms/ResendInviteModal' }

export const Overview = () => (
  <ComponentDocs
    name="ResendInviteModal"
    tier="organism"
    summary="A confirmation before sending another email for one pending invite."
    use={['Open it from a pending invite row before resending email.']}
    avoid={[
      'Do not resend silently from the row menu.',
      'Do not use it for active members.',
    ]}
    rules={[
      'The invitee email is named before submission.',
      'Mutation errors stay in the modal.',
    ]}
    props={[
      { name: 'email', type: 'string', description: 'Invitee being emailed.' },
      {
        name: 'onSubmit',
        type: '() => void',
        description: 'Resends the invitation.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Locks the action while email is being sent.',
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
        <ResendInviteModal email="invitee@example.com" onSubmit={() => {}} />
      )
    }
  />
)
