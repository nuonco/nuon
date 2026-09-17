import { ComponentDocs } from '../__stories__/ComponentDocs'
import { UserProfile } from './UserProfile'

export default {
  title: 'lite/molecules/UserProfile',
}

const USER = {
  name: 'Alex Morgan',
  email: 'alex@example.com',
}

export const Overview = () => (
  <ComponentDocs
    name="UserProfile"
    tier="molecule"
    summary="An avatar with a user name and email address."
    use={[
      'Use as the visible identity inside account controls.',
      'Use compact mode where the surrounding control supplies the accessible label.',
    ]}
    avoid={[
      'Do not add account actions or dropdown state to this component.',
      'Do not assume every account has a name, email, or image.',
    ]}
    rules={[
      'The name falls back to the email and then to User.',
      'Long identity values truncate without shrinking the avatar.',
      'Loading preserves the final profile geometry.',
    ]}
    props={[
      {
        name: 'user',
        type: 'IUserProfileData | null',
        description: 'User identity and optional profile image.',
      },
      {
        name: 'compact',
        type: 'boolean',
        default: 'false',
        description: 'Shows only the avatar.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows avatar and text loading shapes.',
      },
      {
        name: 'avatarSize',
        type: 'TAvatarSize',
        default: "'md'",
        description: 'Size of the profile avatar.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <UserProfile user={USER} />
  </div>
)

export const Compact = () => (
  <div className="p-8">
    <UserProfile user={USER} compact />
  </div>
)

export const MissingName = () => (
  <div className="p-8">
    <UserProfile user={{ email: 'alex@example.com' }} />
  </div>
)

export const MissingIdentity = () => (
  <div className="p-8">
    <UserProfile user={null} />
  </div>
)

export const LongIdentity = () => (
  <div className="w-64 p-8">
    <UserProfile
      user={{
        name: 'Alexandra Morgan-Sanchez with a long name',
        email: 'alexandra.morgan-sanchez@example.com',
      }}
    />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <UserProfile loading />
  </div>
)
