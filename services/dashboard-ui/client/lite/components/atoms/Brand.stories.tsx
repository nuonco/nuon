import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Brand, type TBrandVariant } from './Brand'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Brand',
}

const VARIANTS: TBrandVariant[] = [
  'AWS',
  'Azure',
  'GCP',
  'Docker',
  'OCI',
  'Helm',
  'Terraform',
  'Pulumi',
  'Kubernetes',
  'Lambda',
  'GitHub',
  'Slack',
  'Nuon',
]

const Grid = ({ tone }: { tone?: 'color' | 'mono' }) => (
  <div className="grid grid-cols-[repeat(auto-fit,minmax(9rem,1fr))] gap-3">
    {VARIANTS.map((variant) => (
      <div
        key={variant}
        className="flex min-w-0 items-center gap-3 rounded-lg bg-surface-02 p-3"
      >
        <Brand variant={variant} tone={tone} size={24} className="shrink-0" />
        <Text variant="caption" className="min-w-0">
          {variant}
        </Text>
      </div>
    ))}
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="Brand"
    tier="atom"
    summary="Local SVG brand marks without a runtime icon-package dependency."
    use={[
      'Render cloud, component, connection, and Nuon glyphs.',
      'Use color by default and mono where surrounding text owns the tone.',
    ]}
    avoid={[
      'Do not add brand marks to Icon.',
      'Do not use Brand as the Nuon wordmark or home link.',
      'Do not import react-icons in lite.',
    ]}
    rules={[
      'Brand marks are decorative and always hidden from assistive technology.',
      'Every SVG path lives in the lite source tree.',
      'Nuon is the glyph only; Logo remains the separate wordmark component.',
    ]}
    props={[
      {
        name: 'variant',
        type: 'TBrandVariant',
        description: 'Brand mark to render.',
      },
      {
        name: 'tone',
        type: "'color' | 'mono'",
        default: "'color'",
        description: 'Official colors or currentColor.',
      },
      {
        name: 'size',
        type: 'number | string',
        default: '16',
        description: 'Rendered width and height.',
      },
    ]}
  />
)

export const Color = () => (
  <div className="max-w-4xl p-8">
    <Grid />
  </div>
)

export const Mono = () => (
  <div className="max-w-4xl p-8 text-secondary">
    <Grid tone="mono" />
  </div>
)

export const Sizes = () => (
  <div className="flex items-end gap-6 p-8">
    {[12, 16, 20, 24, 32, 48].map((size) => (
      <div key={size} className="flex flex-col items-center gap-2">
        <Brand variant="Nuon" size={size} />
        <Text variant="label" color="tertiary">
          {size}
        </Text>
      </div>
    ))}
  </div>
)

export const ContrastingSurfaces = () => (
  <div className="grid max-w-4xl grid-cols-1 gap-4 p-8 sm:grid-cols-2">
    <div
      className="rounded-xl bg-white p-4 text-[#1d1d1d]"
      style={{ colorScheme: 'light' }}
    >
      <Grid />
    </div>
    <div
      className="rounded-xl bg-[#17151b] p-4 text-white"
      style={{ colorScheme: 'dark' }}
    >
      <Grid />
    </div>
  </div>
)
