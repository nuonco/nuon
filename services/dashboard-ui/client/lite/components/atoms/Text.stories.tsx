import { Text, type TTextVariant, type TTextColor } from './Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'

export default {
  title: 'lite/atoms/Text',
}

const VARIANTS: Array<{ variant: TTextVariant; usage: string }> = [
  { variant: 'display', usage: 'Empty states, onboarding, the one big line on a page' },
  { variant: 'title', usage: 'Page title' },
  { variant: 'heading', usage: 'Section heading, card heading' },
  { variant: 'body', usage: 'Default. Prose, table cells, most everything' },
  { variant: 'caption', usage: 'Supporting text under a heading, timestamps, hints' },
  { variant: 'label', usage: 'Form labels, table headers, metadata keys' },
]

const COLORS: TTextColor[] = [
  'inherit',
  'primary',
  'secondary',
  'tertiary',
  'accent',
  'positive',
]

export const Overview = () => (
  <ComponentDocs
    name="Text"
    tier="atom"
    summary="Every piece of type in lite. Variants name a role in the page, not a size."
    use={[
      'All rendered text — headings, prose, table cells, labels, IDs, timestamps.',
      'Set as to choose the element, so headings are real headings.',
      'Use the mono family for IDs, timestamps, code and anything the user might copy.',
    ]}
    avoid={[
      'Choosing a variant for its size. Pick the role; if none fits, the scale is missing a step.',
      'Do not wrap a link to size it, because links inherit the surrounding type on their own.',
      'Do not reach for text-size or colour classes. Use variant and color instead.',
    ]}
    rules={[
      'The element is always explicit. Nothing is inferred from the variant, so the document structure is whatever you say it is.',
      'Colour defaults to inherit, so plain text picks up the surface it sits on.',
      'When text is the only child of a padded box, set as to a block element. An inline span takes its line box from the container, which makes the padding lopsided.',
    ]}
    props={[
      { name: 'as', type: 'ElementType', default: "'span'", description: 'The rendered element.' },
      { name: 'variant', type: "'display' | 'title' | 'heading' | 'body' | 'caption' | 'label'", default: "'body'", description: 'Role in the page. Sets size, line-height, tracking and a default weight.' },
      { name: 'color', type: "'inherit' | 'primary' | 'secondary' | 'tertiary' | 'accent' | 'positive'", default: "'inherit'", description: 'Semantic colour token.' },
      { name: 'weight', type: "'normal' | 'medium' | 'semibold'", description: "Overrides the variant's default weight." },
      { name: 'family', type: "'sans' | 'mono'", default: "'sans'", description: 'Typeface.' },
      { name: 'lines', type: 'number', description: 'Clamps to n lines. Also sets the number of skeleton rows while loading.' },
      { name: 'loading', type: 'boolean', default: 'false', description: 'Renders a shimmer in place of the text.' },
      { name: 'loadingWidth', type: 'number', description: 'Skeleton width in ch. Defaults per variant.' },
    ]}
    sections={[
      {
        heading: 'Loading',
        body: 'The skeleton occupies exactly the box the loaded text will, so nothing moves when data arrives. Set loadingWidth to roughly the expected character count; lines gives a multi-row paragraph skeleton.',
      },
    ]}
  />
)

export const Scale = () => (
  <div className="flex flex-col gap-8 p-8">
    {VARIANTS.map(({ variant, usage }) => (
      <div key={variant} className="flex flex-col gap-1">
        <Text variant="label" color="tertiary" family="mono">
          {variant}
        </Text>
        <Text variant={variant}>The quick brown fox jumps over the lazy dog</Text>
        <Text variant="caption" color="tertiary">
          {usage}
        </Text>
      </div>
    ))}
  </div>
)

export const Colors = () => (
  <div className="flex flex-col gap-3 p-8">
    {COLORS.map((color) => (
      <Text key={color} color={color}>
        {color} — the quick brown fox jumps over the lazy dog
      </Text>
    ))}
  </div>
)

export const Weights = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text weight="normal">Normal — the quick brown fox</Text>
    <Text weight="medium">Medium — the quick brown fox</Text>
    <Text weight="semibold">Semibold — the quick brown fox</Text>
  </div>
)

export const Mono = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text family="mono" variant="body">
      inst_01h9k2m4p6q8r0s2t4v6w8x0
    </Text>
    <Text family="mono" variant="caption" color="tertiary">
      2026-08-31T14:27:00Z
    </Text>
    <Text family="mono" variant="label" color="tertiary">
      COMPONENT_ID
    </Text>
  </div>
)

export const Elements = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text as="h1" variant="title">
      A real h1, styled as title
    </Text>
    <Text as="p" variant="body" color="secondary">
      A real p. The element is always explicit — Text never guesses a tag from
      the variant, so the heading level in the document is whatever you say it
      is.
    </Text>
    <Text as="label" variant="label" color="tertiary">
      A real label
    </Text>
  </div>
)

export const Clamped = () => (
  <div className="max-w-md p-8">
    <Text as="p" lines={2} color="secondary">
      Nuon deploys your application into your customers cloud accounts, so this
      paragraph is long enough to demonstrate that clamping to two lines works
      and truncates with an ellipsis rather than overflowing its container or
      pushing the layout around.
    </Text>
  </div>
)

export const Loading = () => (
  <div className="flex flex-col gap-8 p-8">
    {VARIANTS.map(({ variant }) => (
      <div key={variant} className="flex flex-col gap-1">
        <Text variant="label" color="tertiary" family="mono">
          {variant}
        </Text>
        <Text variant={variant} loading />
      </div>
    ))}
  </div>
)

export const LoadingParagraph = () => (
  <div className="max-w-md p-8">
    <Text as="p" lines={3} loading />
  </div>
)

export const LoadingDoesNotShiftLayout = () => (
  <div className="flex flex-col gap-8 p-8">
    <Text variant="caption" color="tertiary">
      The same component loading and loaded. The boxes line up to the pixel, so
      nothing moves when the data arrives.
    </Text>
    {VARIANTS.map(({ variant }) => (
      <div key={variant} className="flex items-baseline gap-6">
        <Text variant={variant} loading />
        <Text variant={variant}>Loaded {variant}</Text>
      </div>
    ))}
  </div>
)

export const InContext = () => (
  <div className="flex max-w-lg flex-col gap-6 p-8">
    <div className="flex flex-col gap-1">
      <Text as="h1" variant="title">
        Installs
      </Text>
      <Text variant="caption" color="tertiary">
        Every deployment of this app into a customer account.
      </Text>
    </div>

    <div className="flex flex-col gap-3 rounded-xl border border-divider bg-surface-01 p-5">
      <div className="flex flex-col gap-1">
        <Text as="h2" variant="heading">
          acme-production
        </Text>
        <Text variant="caption" family="mono" color="tertiary">
          inst_01h9k2m4p6q8r0s2t4v6w8x0
        </Text>
      </div>
      <div className="flex flex-col gap-1">
        <Text variant="label" color="tertiary">
          Region
        </Text>
        <Text variant="body">us-west-2</Text>
      </div>
    </div>
  </div>
)
