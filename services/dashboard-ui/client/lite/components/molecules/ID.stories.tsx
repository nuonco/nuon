import { ComponentDocs } from '../__stories__/ComponentDocs'
import { ID } from './ID'
import { Link } from '../atoms/Link'
import { Text } from '../atoms/Text'

export default {
  title: 'lite/molecules/ID',
}

const INSTALL = 'inst_01h9k2m4p6q8r0s2t4v6w8x0'

export const Overview = () => (
  <ComponentDocs
    name="ID"
    tier="molecule"
    summary="A resource identifier in mono, with a copy affordance."
    use={[
      'Show any opaque identifier the user might paste into a CLI, a ticket or a support thread.',
      'Truncate in tables and tight rows, where the full identifier would dominate.',
    ]}
    avoid={[
      'Do not use it for human-readable names. Those are Text, or a Link if they navigate.',
      'Never make the identifier itself a link. Navigation belongs on the resource name above it, or on a display name such as "Component build" where the resource has no name of its own.',
      'Do not truncate with copying turned off. A shortened identifier the user cannot copy in full is worse than none.',
    ]}
    rules={[
      'Copy always writes the full value, even when the displayed text is truncated.',
      'Truncation elides the middle, so both the prefix and the distinguishing tail survive.',
      'The copy button is the only hit target, and the text is selectable rather than clickable.',
    ]}
    props={[
      { name: 'value', type: 'string', description: 'The full identifier. Always what gets copied.' },
      { name: 'truncate', type: 'boolean', default: 'false', description: 'Shortens the displayed text with a middle ellipsis.' },
      { name: 'head', type: 'number', default: '10', description: 'Characters kept at the start when truncating.' },
      { name: 'tail', type: 'number', default: '4', description: 'Characters kept at the end when truncating.' },
      { name: 'copyable', type: 'boolean', default: 'true', description: 'Set false where a copy button would be noise.' },
      { name: 'label', type: 'string', default: "'Copy ID'", description: 'Accessible name for the copy button.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Shimmer, delegated to Text.' },
      { name: 'loadingWidth', type: 'number', description: 'Skeleton width in ch.' },
    ]}
  />
)

export const Default = () => (
  <div className="p-8">
    <ID value={INSTALL} label="Copy install ID" />
  </div>
)

export const Truncated = () => (
  <div className="flex flex-col gap-3 p-8">
    <ID value={INSTALL} truncate />
    <Text variant="caption" color="tertiary">
      Both ends survive; only the middle is elided. Copy still writes the full
      value.
    </Text>
  </div>
)

export const WithoutCopy = () => (
  <div className="p-8">
    <ID value={INSTALL} copyable={false} />
  </div>
)

export const UnderAResourceName = () => (
  <div className="flex flex-col gap-6 p-8">
    <div className="flex flex-col gap-0.5">
      <Link href="/org123/installs/inst456" variant="heading">
        acme-production
      </Link>
      <ID value={INSTALL} />
    </div>
    <div className="flex flex-col gap-0.5">
      <Link href="/org123/builds/bld789" variant="heading">
        Component build
      </Link>
      <ID value="bld_01k4m6p8r0t2v4x6z8b0d2f4" label="Copy build ID" />
    </div>
    <Text variant="caption" color="tertiary">
      The name carries the link. Where a resource has no name of its own, a
      display name does the job.
    </Text>
  </div>
)

export const InATable = () => (
  <div className="max-w-xl p-8">
    <div className="flex flex-col divide-y divide-divider rounded-xl border border-divider">
      {[
        ['acme-production', 'inst_01h9k2m4p6q8r0s2t4v6w8x0'],
        ['acme-staging', 'inst_02k4m6p8r0t2v4x6z8b0d2f4'],
        ['payments-eu', 'inst_03m6p8r0t2v4x6z8b0d2f4h6'],
      ].map(([name, id]) => (
        <div key={id} className="flex items-center justify-between gap-4 px-4 py-2">
          <Text variant="body">{name}</Text>
          <ID value={id} truncate />
        </div>
      ))}
    </div>
  </div>
)

export const Loading = () => (
  <div className="flex flex-col gap-2 p-8">
    <ID value={INSTALL} loading />
    <ID value={INSTALL} />
  </div>
)
