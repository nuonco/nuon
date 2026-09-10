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
      'Pass one short metadata value when it helps distinguish similar items.',
    ]}
    avoid={[
      'Do not use it for navigation to a page. This button opens a panel.',
      'Do not put multiple metadata blocks or actions in the row.',
      'Do not make ID copyable inside the row because nested buttons are invalid.',
    ]}
    rules={[
      'The whole row is one real button with a visible focus ring.',
      'Name truncates before metadata, ID, and the panel chevron.',
      'Loading preserves the same icon, row height, and chevron.',
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
        name: 'metadata',
        type: 'ReactNode',
        description: 'One short supporting value before the ID.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads the name and metadata within the real row.',
      },
    ]}
  />
)

export const Install = () => (
  <div className="max-w-2xl p-8">
    <ConfigItem
      icon="ShippingContainerIcon"
      name="payments-production"
      id="inst_01h9k2m4p6r8t0v2"
      metadata="AWS"
    />
  </div>
)

export const LongName = () => (
  <div className="max-w-md p-8">
    <ConfigItem
      name="a-very-long-configured-resource-name-that-must-not-push-out-the-id"
      id="cmp_01h9k2m4p6r8t0v2"
      metadata="Terraform"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-8">
    <ConfigItem loading />
  </div>
)
