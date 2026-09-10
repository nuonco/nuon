import { Status } from '../atoms/Status'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { ConfigItem } from './ConfigItem'

export default {
  title: 'lite/molecules/ConfigItem',
}

export const Overview = () => (
  <ComponentDocs
    name="ConfigItem"
    tier="molecule"
    summary="A compact configuration or resource line that opens a detail panel."
    use={[
      'Show one configured resource or resolved install inside a compact card list.',
      'Pass status facets on the top line and one short supporting value on the second.',
    ]}
    avoid={[
      'Do not use it for navigation to a page. This button opens a panel.',
      'Do not put actions in the row. It is a single button.',
      'Do not make ID copyable inside the row because nested buttons are invalid.',
    ]}
    rules={[
      'The whole row is one real button with a visible focus ring.',
      'Two lines always: name and status on top, ID and metadata below. It does not collapse to one line at any width.',
      'Name truncates before status, and the second line wraps rather than overflowing.',
      'Loading preserves the same icon, both lines, and the chevron.',
    ]}
    props={[
      {
        name: 'icon',
        type: 'TIconVariant',
        default: "'CubeIcon'",
        description: 'Type icon at the start of the row.',
      },
      {
        name: 'name',
        type: 'string',
        description: 'Primary resource name.',
      },
      {
        name: 'id',
        type: 'string',
        description: 'Resource ID, middle-truncated and not copyable in the button.',
      },
      {
        name: 'status',
        type: 'ReactNode',
        description: 'Status facets shown beside the name on the top line.',
      },
      {
        name: 'metadata',
        type: 'ReactNode',
        description: 'One short supporting value after the ID on the second line.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads both lines within the real row.',
      },
    ]}
  />
)

export const Install = () => (
  <div className="max-w-2xl p-8">
    <ConfigItem
      name="payments-production"
      id="inst_01h9k2m4p6r8t0v2"
      status={<Status status="active" variant="icon" />}
      metadata="US East (N. Virginia)"
    />
  </div>
)

export const Narrow = () => (
  <div className="max-w-xs p-8">
    <ConfigItem
      name="payments-production"
      id="inst_01h9k2m4p6r8t0v2"
      status={<Status status="active" variant="icon" />}
      metadata="US East (N. Virginia)"
    />
  </div>
)

export const LongName = () => (
  <div className="max-w-md p-8">
    <ConfigItem
      name="a-very-long-configured-resource-name-that-must-not-push-out-the-id"
      id="cmp_01h9k2m4p6r8t0v2"
      status={<Status status="active" variant="icon" />}
      metadata="Terraform"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-8">
    <ConfigItem loading />
  </div>
)
