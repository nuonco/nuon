import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { ChangeRoleModal } from './ChangeRoleModal'

export default { title: 'lite/organisms/ChangeRoleModal' }

const roles = [
  {
    id: 'role_admin',
    role_type: 'org_admin',
    title: 'Org admin',
    description: 'Full access to org settings and resources.',
  },
  {
    id: 'role_read',
    role_type: 'org_read_only',
    title: 'Read only',
    description: 'View access to org resources.',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="ChangeRoleModal"
    tier="organism"
    summary="A focused role change for one active team member."
    use={['Open it from an active member row when the caller is an org admin.']}
    avoid={[
      'Do not offer it on the current account row or to non-admins.',
      'Do not use it for pending invites.',
    ]}
    rules={[
      'Role options show their configured titles and descriptions.',
      'The save action remains disabled until the role changes.',
      'Mutation errors stay in the modal.',
    ]}
    props={[
      { name: 'email', type: 'string', description: 'Member being updated.' },
      {
        name: 'roles',
        type: 'TRoleInfo[]',
        description: 'Roles available for team members.',
      },
      {
        name: 'currentRole',
        type: 'string',
        description: 'Role currently assigned to the member.',
      },
      {
        name: 'onSubmit',
        type: '(roleType: string) => void',
        description: 'Saves the selected role.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Locks the action while the role is being saved.',
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
        <ChangeRoleModal
          email="member@example.com"
          roles={roles}
          currentRole="org_read_only"
          onSubmit={() => {}}
        />
      )
    }
  />
)

export const Pending = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <ChangeRoleModal
          email="member@example.com"
          roles={roles}
          currentRole="org_read_only"
          onSubmit={() => {}}
          pending
        />
      )
    }
  />
)
