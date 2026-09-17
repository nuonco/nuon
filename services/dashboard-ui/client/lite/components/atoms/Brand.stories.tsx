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
      'size is the optical size, not a box: marks are normalized to equal visual area, so a wide mark renders shorter and wider than a compact one.',
      'Height never exceeds size, so a mark always fits the line it sits on; width varies with the artwork.',
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
        description:
          'Optical size. Caps the height and scales the width to the mark\u2019s aspect ratio. Accepts any CSS length, including em.',
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

export const OpticalBalance = () => (
  <div className="flex flex-col gap-6 p-8">
    <Text variant="caption" color="tertiary">
      All marks at size 24, on a shared baseline.
    </Text>
    <div className="flex flex-wrap items-baseline gap-6">
      {VARIANTS.map((variant) => (
        <span key={variant} className="flex items-baseline gap-2">
          <Brand variant={variant} size={24} />
          <Text variant="caption">{variant}</Text>
        </span>
      ))}
    </div>
    <div className="flex flex-wrap items-center gap-6">
      {VARIANTS.map((variant) => (
        <span
          key={variant}
          className="flex items-center gap-2 bg-surface-02 p-2"
        >
          <Brand variant={variant} size={24} />
        </span>
      ))}
    </div>
  </div>
)

export const InlineWithText = () => (
  <div className="flex max-w-md flex-col gap-3 p-8">
    {VARIANTS.slice(0, 5).map((variant) => (
      <Text key={variant} variant="caption" className="flex items-center gap-2">
        <Brand variant={variant} size="1.25em" />
        <span>Runs on {variant}</span>
      </Text>
    ))}
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
