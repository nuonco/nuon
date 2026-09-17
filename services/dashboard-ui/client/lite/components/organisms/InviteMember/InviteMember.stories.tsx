import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { InviteMember } from './InviteMember'

export default { title: 'lite/organisms/InviteMember' }

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
    name="InviteMember"
    tier="organism"
    summary="A modal form that invites one person to the org with a selected role."
    use={[
      'Open it from the Team page as the page-level create action.',
      'Render the container so roles, mutation state, invalidation, and feedback remain encapsulated.',
    ]}
    avoid={[
      'Do not place the fields directly on the Team page.',
      'Do not show submission failures in a toast.',
      'Do not add native required attributes alongside Zod validation.',
    ]}
    rules={[
      'Email and role remain flat TanStack Form fields.',
      'Zod is the only validation source.',
      'The primary action remains disabled until both fields are valid.',
      'Server errors remain in FormErrorBanner inside the form.',
    ]}
    props={[
      {
        name: 'roles',
        type: 'TRoleInfo[]',
        description: 'Team roles shown with titles and descriptions.',
      },
      {
        name: 'onSubmit',
        type: '(values: TInviteMemberValues) => void',
        description: 'Submits the validated email and role.',
      },
      {
        name: 'pending',
        type: 'boolean',
        default: 'false',
        description: 'Locks the action while the invitation is being sent.',
      },
      {
        name: 'rolesLoading',
        type: 'boolean',
        default: 'false',
        description: 'Shows the role control loading state.',
      },
      {
        name: 'error',
        type: 'TAPIError | Error | null',
        description: 'Displays the server failure inside the form.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(<InviteMember roles={roles} onSubmit={() => {}} />)
    }
  />
)

export const Submitting = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(<InviteMember roles={roles} onSubmit={() => {}} pending />)
    }
  />
)

export const ServerError = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(
        <InviteMember
          roles={roles}
          onSubmit={() => {}}
          error={{
            error: 'Invite failed',
            description: 'An invitation already exists for this email.',
            user_error: true,
          }}
        />
      )
    }
  />
)

export const Validation = () => (
  <SurfaceStory
    open={({ openModal }) =>
      openModal(<InviteMember roles={roles} onSubmit={() => {}} />)
    }
  />
)
