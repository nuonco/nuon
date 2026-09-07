import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { UserDropdown } from './UserDropdown'

export default {
  title: 'lite/organisms/UserDropdown',
}

const USER = {
  name: 'Alex Morgan',
  email: 'alex@example.com',
}

const SIGN_OUT_HREF = 'https://auth.example.com/logout'

export const Overview = () => (
  <ComponentDocs
    name="UserDropdown"
    tier="organism"
    summary="A user-profile trigger with account actions."
    use={[
      'Use in application chrome where the signed-in user needs account actions.',
      'Use compact mode for an icon-sized header control.',
    ]}
    avoid={[
      'Do not add organization navigation or setup actions to this menu.',
      'Do not open sign-out in a new browser tab.',
    ]}
    rules={[
      'The visible trigger is always UserProfile.',
      'Sign out is the only menu item until the account menu is designed.',
      'The sign-out destination performs a same-window navigation.',
    ]}
    props={[
      {
        name: 'user',
        type: 'IUserProfileData | null',
        description: 'Identity rendered in the trigger.',
      },
      {
        name: 'signOutHref',
        type: 'string',
        description: 'Authentication-service logout URL.',
      },
      {
        name: 'compact',
        type: 'boolean',
        default: 'false',
        description: 'Uses an avatar-only trigger.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows profile loading shapes in the trigger.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="flex justify-end p-20">
    <UserDropdown user={USER} signOutHref={SIGN_OUT_HREF} defaultOpen />
  </div>
)

export const Compact = () => (
  <div className="flex justify-end p-20">
    <UserDropdown user={USER} signOutHref={SIGN_OUT_HREF} compact defaultOpen />
  </div>
)

export const Loading = () => (
  <div className="flex justify-end p-20">
    <UserDropdown loading signOutHref={SIGN_OUT_HREF} defaultOpen />
  </div>
)

export const LongIdentity = () => (
  <div className="flex justify-end p-20">
    <UserDropdown
      user={{
        name: 'Alexandra Morgan-Sanchez with a long name',
        email: 'alexandra.morgan-sanchez@example.com',
      }}
      signOutHref={SIGN_OUT_HREF}
      defaultOpen
    />
  </div>
)

export const SidebarPlacement = () => (
  <div className="flex min-h-[28rem] items-end p-8">
    <Card
      as="aside"
      padding="sm"
      className="flex h-96 w-64 flex-col justify-between"
    >
      <Text variant="caption" color="secondary">
        Sidebar content
      </Text>
      <UserDropdown
        user={USER}
        signOutHref={SIGN_OUT_HREF}
        side="top"
        align="start"
        stretch
        defaultOpen
      />
    </Card>
  </div>
)
