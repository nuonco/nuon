import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Link } from './Link'
import { Logo } from './Logo'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Logo',
}

export const Overview = () => (
  <ComponentDocs
    name="Logo"
    tier="atom"
    summary="The Nuon product wordmark for application chrome."
    use={[
      'Render the full wordmark in focused headers and expanded navigation.',
      'Render the mark in compact or collapsed application chrome.',
    ]}
    avoid={[
      'Do not use Logo as a generic vendor mark.',
      'Do not put navigation behavior inside the artwork component.',
      'Do not use Brand variant Nuon when the product wordmark is intended.',
    ]}
    rules={[
      'Color uses the gradient mark; mono makes the entire logo inherit currentColor.',
      'Logo is decorative artwork; an interactive wrapper provides its accessible name.',
      'Use the same local SVG geometry in every theme.',
    ]}
    props={[
      {
        name: 'variant',
        type: "'wordmark' | 'mark'",
        default: "'wordmark'",
        description: 'Full Nuon wordmark or compact mark.',
      },
      {
        name: 'size',
        type: 'number | string',
        default: '28',
        description: 'Rendered artwork height.',
      },
      {
        name: 'tone',
        type: "'color' | 'mono'",
        default: "'color'",
        description: 'Gradient mark or currentColor artwork.',
      },
    ]}
  />
)

export const Wordmark = () => (
  <div className="p-8">
    <Logo />
  </div>
)

export const Mark = () => (
  <div className="p-8">
    <Logo variant="mark" />
  </div>
)

export const Mono = () => (
  <div className="flex items-center gap-8 p-8 text-secondary">
    <Logo tone="mono" />
    <Logo variant="mark" tone="mono" />
  </div>
)

export const Sizes = () => (
  <div className="flex flex-col items-start gap-6 p-8">
    {[20, 28, 36, 48].map((size) => (
      <div key={size} className="flex items-center gap-4">
        <Logo size={size} />
        <Text variant="label" color="tertiary">
          {size}px
        </Text>
      </div>
    ))}
  </div>
)

export const Surfaces = () => (
  <div className="grid max-w-3xl gap-4 p-8 sm:grid-cols-2">
    <div
      className="rounded-xl bg-white p-6 text-[#1b242c]"
      style={{ colorScheme: 'light' }}
    >
      <Logo />
    </div>
    <div
      className="rounded-xl bg-[#17151b] p-6 text-white"
      style={{ colorScheme: 'dark' }}
    >
      <Logo />
    </div>
  </div>
)

export const Navigation = () => (
  <div className="p-8">
    <Link
      href="/"
      aria-label="Nuon dashboard"
      style={{ color: 'var(--text-primary)' }}
    >
      <Logo />
    </Link>
  </div>
)
