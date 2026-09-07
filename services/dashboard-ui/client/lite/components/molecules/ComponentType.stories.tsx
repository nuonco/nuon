import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Text } from '../atoms/Text'
import { ComponentType, type TComponentTypeValue } from './ComponentType'

export default {
  title: 'lite/molecules/ComponentType',
}

const TYPES: TComponentTypeValue[] = [
  'docker_build',
  'external_image',
  'helm_chart',
  'terraform_module',
  'job',
  'kubernetes_manifest',
  'pulumi',
  'unknown',
]

export const Overview = () => (
  <ComponentDocs
    name="ComponentType"
    tier="molecule"
    summary="A component type mark and label backed by local brand artwork."
    use={[
      'Identify component types in tables, cards, filters, and metadata.',
      'Use icon display only where space is constrained.',
    ]}
    avoid={[
      'Do not import component brand marks directly into feature UI.',
      'Do not use an unrelated generic icon for a known component type.',
    ]}
    rules={[
      'Color is the default.',
      'Unknown values render a question mark and safe label.',
      'Icon-only display includes a keyboard-accessible tooltip.',
    ]}
    props={[
      {
        name: 'type',
        type: 'TComponentTypeValue',
        description: 'Component type to display.',
      },
      {
        name: 'display',
        type: "'abbr' | 'name' | 'icon'",
        default: "'name'",
        description: 'Label form or icon-only display.',
      },
      {
        name: 'tone',
        type: "'color' | 'mono'",
        default: "'color'",
        description: 'Brand colors or currentColor.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="grid max-w-2xl grid-cols-2 gap-5 p-8">
    {TYPES.map((type) => (
      <ComponentType key={type} type={type} />
    ))}
  </div>
)

export const Abbreviations = () => (
  <div className="flex flex-wrap items-center gap-6 p-8">
    {TYPES.map((type) => (
      <ComponentType key={type} type={type} display="abbr" />
    ))}
  </div>
)

export const Icons = () => (
  <div className="flex items-center gap-6 p-8">
    {TYPES.map((type) => (
      <ComponentType key={type} type={type} display="icon" iconSize={24} />
    ))}
  </div>
)

export const Mono = () => (
  <div className="grid max-w-2xl grid-cols-2 gap-5 p-8 text-secondary">
    {TYPES.map((type) => (
      <ComponentType key={type} type={type} tone="mono" />
    ))}
  </div>
)

export const Sizes = () => (
  <div className="flex flex-col gap-5 p-8">
    <ComponentType type="docker_build" variant="caption" iconSize={14} />
    <ComponentType type="helm_chart" variant="body" iconSize={18} />
    <ComponentType
      type="terraform_module"
      variant="heading"
      weight="semibold"
      iconSize={26}
    />
  </div>
)

export const Inline = () => (
  <Text className="p-8">
    The <ComponentType type="kubernetes_manifest" variant="caption" /> component
    is ready to deploy.
  </Text>
)
