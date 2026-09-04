import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Avatar, type TAvatarSize } from './Avatar'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Avatar',
}

const PROFILE_IMAGE =
  'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"%3E%3Crect width="64" height="64" fill="%234cc9f0"/%3E%3Ccircle cx="32" cy="25" r="12" fill="%23141217"/%3E%3Cpath d="M10 64c2-15 10-23 22-23s20 8 22 23" fill="%23141217"/%3E%3C/svg%3E'

const SIZES: TAvatarSize[] = ['xs', 'sm', 'md', 'lg', 'xl']

export const Overview = () => (
  <ComponentDocs
    name="Avatar"
    tier="atom"
    summary="A user image with initials and icon fallbacks."
    use={[
      'Identify a user in profile controls, tables, and activity.',
      'Provide a name even when an image URL is available so fallback content remains useful.',
    ]}
    avoid={[
      'Do not use an avatar as the accessible name of an interactive control.',
      'Do not assume remote profile images will load.',
    ]}
    rules={[
      'Images fall back to initials after loading errors.',
      'A missing name and image falls back to the user icon.',
      'Pass alt only when the avatar is not accompanied by visible user text.',
    ]}
    props={[
      {
        name: 'name',
        type: 'string',
        description: 'User name used to produce fallback initials.',
      },
      {
        name: 'src',
        type: 'string',
        description: 'Optional profile image URL.',
      },
      {
        name: 'size',
        type: "'xs' | 'sm' | 'md' | 'lg' | 'xl'",
        default: "'md'",
        description: 'Avatar dimensions and fallback size.',
      },
      {
        name: 'shape',
        type: "'circle' | 'rounded'",
        default: "'circle'",
        description: 'Circular or softly rounded artwork.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows the avatar loading shape.',
      },
    ]}
  />
)

export const Initials = () => (
  <div className="flex items-center gap-4 p-8">
    <Avatar name="Alex Morgan" />
    <Avatar name="Priya" />
    <Avatar name="Lee Chen" />
    <Avatar />
  </div>
)

export const Image = () => (
  <div className="flex items-center gap-4 p-8">
    <Avatar name="Alex Morgan" src={PROFILE_IMAGE} />
    <Avatar name="Alex Morgan" src={PROFILE_IMAGE} size="lg" />
    <Avatar name="Alex Morgan" src={PROFILE_IMAGE} size="xl" />
  </div>
)

export const BrokenImage = () => (
  <div className="flex items-center gap-4 p-8">
    <Avatar name="Alex Morgan" src="/missing-avatar.png" />
    <Text variant="caption" color="secondary">
      The missing image falls back to initials.
    </Text>
  </div>
)

export const Sizes = () => (
  <div className="flex items-end gap-6 p-8">
    {SIZES.map((size) => (
      <div key={size} className="flex flex-col items-center gap-2">
        <Avatar name="Alex Morgan" size={size} />
        <Text variant="label" color="tertiary">
          {size}
        </Text>
      </div>
    ))}
  </div>
)

export const Shapes = () => (
  <div className="flex items-center gap-4 p-8">
    <Avatar name="Alex Morgan" shape="circle" size="lg" />
    <Avatar name="Alex Morgan" shape="rounded" size="lg" />
  </div>
)

export const Loading = () => (
  <div className="flex items-center gap-4 p-8">
    {SIZES.map((size) => (
      <Avatar key={size} size={size} loading />
    ))}
  </div>
)
