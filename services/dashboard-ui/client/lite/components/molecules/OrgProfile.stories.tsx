import { ComponentDocs } from '../__stories__/ComponentDocs'
import { OrgProfile } from './OrgProfile'

export default {
  title: 'lite/molecules/OrgProfile',
}

const LOGO =
  'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"%3E%3Crect width="64" height="64" rx="14" fill="%236366f1"/%3E%3Cpath d="M18 42V22h8l12 12V22h8v20h-8L26 30v12z" fill="white"/%3E%3C/svg%3E'

const ORG = {
  id: 'org_01k4m6p8r0t2v4x6z8b0d2f4',
  name: 'Acme',
  status: 'active',
}

export const Overview = () => (
  <ComponentDocs
    name="OrgProfile"
    tier="molecule"
    summary="An organization logo with its name, status, and ID."
    use={[
      'Use as the visible organization identity in context controls and application chrome.',
      'Use the loading state wherever the organization is fetched asynchronously.',
    ]}
    avoid={[
      'Do not add organization actions or switcher state to this component.',
      'Do not place interactive controls inside the profile.',
    ]}
    rules={[
      'The logo falls back to initials derived from the organization name.',
      'The status dot sits to the left of the name, with the organization ID below it.',
      'Loading preserves the final profile geometry.',
    ]}
    props={[
      {
        name: 'org',
        type: 'TOrg | null',
        description: 'Organization identity and status data.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Shows logo and text loading shapes.',
      },
      {
        name: 'avatarSize',
        type: 'TAvatarSize',
        default: "'md'",
        description: 'Size of the organization logo.',
      },
      {
        name: 'variant',
        type: "'full' | 'inline'",
        default: "'full'",
        description:
          'Full shows the logo, status, name and ID. Inline shows only the status dot and name.',
      },
    ]}
  />
)

export const WithLogo = () => (
  <div className="p-8">
    <OrgProfile org={{ ...ORG, logo_url: LOGO }} />
  </div>
)

export const InitialsFallback = () => (
  <div className="p-8">
    <OrgProfile org={ORG} />
  </div>
)

export const Inline = () => (
  <div className="p-8">
    <OrgProfile org={ORG} variant="inline" />
  </div>
)

export const InlineLoading = () => (
  <div className="p-8">
    <OrgProfile loading variant="inline" />
  </div>
)

export const LongIdentity = () => (
  <div className="w-72 p-8">
    <OrgProfile
      org={{
        id: 'org_01k4m6p8r0t2v4x6z8b0d2f4',
        name: 'Example organization with a long name',
        status: 'active',
      }}
    />
  </div>
)

export const MissingData = () => (
  <div className="p-8">
    <OrgProfile org={null} />
  </div>
)

export const Loading = () => (
  <div className="p-8">
    <OrgProfile loading />
  </div>
)
