import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Input } from './Input'
import { Label } from './Label'

export default { title: 'lite/atoms/Label' }

export const Overview = () => (
  <ComponentDocs
    name="Label"
    tier="atom"
    summary="The semantic label for a form control, sized with the lite type scale."
    use={[
      'Use through Field for ordinary controls.',
      'Use directly for custom control layouts.',
    ]}
    avoid={['Do not use plain Text when a visible string labels a control.']}
    rules={[
      'Connect it to its control with htmlFor.',
      'Required state comes from Zod, not an asterisk or native required.',
    ]}
    props={[
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Renders the label-shaped loading state.',
      },
      {
        name: 'loadingWidth',
        type: 'number',
        description: 'Loading width in ch.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-sm p-8">
    <Label htmlFor="label-default">Install name</Label>
    <Input id="label-default" className="mt-1.5" />
  </div>
)

export const Loading = () => (
  <div className="max-w-sm p-8">
    <Label loading loadingWidth={12}>
      Install name
    </Label>
    <Input loading className="mt-1.5" />
  </div>
)
