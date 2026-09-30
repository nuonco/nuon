import type { ReactNode } from 'react'
import type { TAPIError } from '@/types'
import { UpdateOrgNameForm } from './UpdateOrgName'

export default { title: 'Features / Orgs / Update name' }

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="max-w-md p-6">{children}</div>
)

export const Unchanged = () => (
  <Frame>
    <UpdateOrgNameForm
      currentName="acme"
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </Frame>
)

export const Saving = () => (
  <Frame>
    <UpdateOrgNameForm
      currentName="acme"
      isPending
      error={null}
      onSubmit={() => {}}
    />
  </Frame>
)

export const NameTaken = () => (
  <Frame>
    <UpdateOrgNameForm
      currentName="acme"
      isPending={false}
      error={
        {
          error: 'An organization with this name already exists.',
          description: 'Choose a different name.',
          user_error: true,
          status: 409,
        } as TAPIError
      }
      onSubmit={() => {}}
    />
  </Frame>
)

export const Forbidden = () => (
  <Frame>
    <UpdateOrgNameForm
      currentName="acme"
      isPending={false}
      error={
        {
          error: 'this action requires write access to organization settings',
          description:
            'Your role (Read-only) does not have write access to organization settings. Ask an organization admin to assign a role that does.',
          user_error: true,
          status: 403,
        } as TAPIError
      }
      onSubmit={() => {}}
    />
  </Frame>
)
