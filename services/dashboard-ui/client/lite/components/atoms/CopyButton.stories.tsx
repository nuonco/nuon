import { ComponentDocs } from '../__stories__/ComponentDocs'
import { CopyButton } from './CopyButton'
import { Text } from './Text'

export default {
  title: 'lite/atoms/CopyButton',
}

export const Overview = () => (
  <ComponentDocs
    name="CopyButton"
    tier="atom"
    summary="Copies an explicit value to the clipboard, reporting the result through its tooltip."
    use={[
      'Place one beside anything worth copying, such as an ID, a token, an ARN, an output or a command.',
      'Embed one in a component that offers copy, where that component takes a value and renders this.',
    ]}
    avoid={[
      'Do not re-implement the clipboard write. Use the useCopy hook when you need copy without a button.',
      'Do not copy something the user cannot see. Copy what is on screen.',
    ]}
    rules={[
      'The value is required and is exactly what gets written. It is never inferred from children.',
      'Feedback comes through the tooltip, where the icon becomes a check and the label becomes "Copied".',
      'Failure is reported rather than swallowed, and the tooltip explains how to copy manually.',
      'The small size sits inline with caption text, and the medium size matches a normal button row.',
    ]}
    props={[
      { name: 'value', type: 'string', description: 'Exact text written to the clipboard.' },
      { name: 'label', type: 'string', default: "'Copy'", description: 'Accessible name and idle tooltip text.' },
      { name: 'size', type: "'md' | 'sm'", default: "'sm'", description: 'Button size.' },
      { name: 'variant', type: "'primary' | 'secondary' | 'ghost' | 'danger'", default: "'ghost'", description: 'Button variant.' },
      { name: 'side', type: "'top' | 'bottom' | 'left' | 'right'", default: "'top'", description: 'Preferred tooltip side.' },
    ]}
    sections={[
      {
        heading: 'useCopy',
        body: 'The hook behind this button owns the write, the reset timer and the failure path, including a fallback for contexts where the clipboard API is unavailable. Use it directly when copy needs a different affordance.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="flex items-center gap-2 p-8">
    <Text variant="caption" family="mono" color="tertiary">
      inst_01h9k2m4p6q8r0s2t4v6w8x0
    </Text>
    <CopyButton value="inst_01h9k2m4p6q8r0s2t4v6w8x0" label="Copy install ID" />
  </div>
)

export const Sizes = () => (
  <div className="flex items-center gap-6 p-8">
    <div className="flex items-center gap-2">
      <CopyButton value="small" size="sm" />
      <Text variant="caption" color="tertiary">
        sm
      </Text>
    </div>
    <div className="flex items-center gap-2">
      <CopyButton value="medium" size="md" />
      <Text variant="caption" color="tertiary">
        md
      </Text>
    </div>
  </div>
)

export const Variants = () => (
  <div className="flex items-center gap-3 p-8">
    <CopyButton value="ghost" variant="ghost" />
    <CopyButton value="secondary" variant="secondary" />
    <CopyButton value="primary" variant="primary" />
  </div>
)

export const EmptyValueFails = () => (
  <div className="flex items-center gap-3 p-8">
    <CopyButton value="" label="Copy nothing" />
    <Text variant="caption" color="tertiary">
      An empty value reports the failure state rather than claiming success.
    </Text>
  </div>
)
