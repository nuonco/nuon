import type { ReactNode } from 'react'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Badge } from './Badge'
import { Button } from './Button'
import { Card } from './Card'
import { Status } from './Status'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Card',
}

export const Overview = () => (
  <ComponentDocs
    name="Card"
    tier="atom"
    summary="A rounded, translucent surface that groups related content and blurs whatever sits behind it."
    use={[
      'Group a small set of related values, such as a metadata block or a summary panel.',
      'Lay a panel over busy content — logs, a diff, a gradient — where the blur keeps text readable.',
      'Nest with padding none when the child owns its own spacing, such as a table or a code block.',
    ]}
    avoid={[
      'Do not nest cards more than one level deep. The tints stack and the hierarchy stops reading.',
      'Do not add a border. The tint and the blur carry the edge in every theme except high contrast, which adds one for you.',
      'Do not use a card as the page background. It is a surface for content, not a layout shell.',
    ]}
    rules={[
      'The background is a translucent tint of the theme foreground, so a card reads the same over any surface.',
      'The high contrast theme swaps the tint for an opaque fill and a solid border, and drops the blur.',
      'interactive is visual only. Wrap the content in a Link or render a real button for the actual affordance.',
    ]}
    props={[
      {
        name: 'padding',
        type: "'none' | 'sm' | 'md' | 'lg'",
        default: "'md'",
        description: 'Inner spacing. Use none when the child brings its own.',
      },
      {
        name: 'blur',
        type: "'none' | 'sm' | 'md' | 'lg'",
        default: "'md'",
        description:
          'Strength of the backdrop blur. Use none for cards over a flat surface.',
      },
      {
        name: 'opacity',
        type: "'default' | 'strong' | 'solid'",
        default: "'default'",
        description:
          'Background opacity. Strong is intended for floating surfaces; solid removes transparency.',
      },
      {
        name: 'interactive',
        type: 'boolean',
        default: 'false',
        description:
          'Adds a hover tint, pointer cursor and focus ring for cards that navigate.',
      },
      {
        name: 'shadow',
        type: "'none' | 'default' | 'floating'",
        default: "'default'",
        description:
          'Surface elevation. Floating is reserved for persistent glass chrome.',
      },
    ]}
  />
)

const Backdrop = ({ children }: { children: ReactNode }) => (
  <div className="relative overflow-hidden rounded-xl p-8">
    <div
      aria-hidden
      className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,#4cc9f0_0%,transparent_55%),radial-gradient(circle_at_80%_10%,#f72585_0%,transparent_50%),radial-gradient(circle_at_50%_90%,#3a00ff_0%,transparent_55%)] opacity-70"
    />
    <div className="relative">{children}</div>
  </div>
)

const Metadata = () => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center justify-between gap-4">
      <Text variant="body" weight="medium">
        payments-api
      </Text>
      <Status status="active" />
    </div>
    <Text variant="caption" color="tertiary">
      Deployed 4 minutes ago to us-west-2 from the release branch.
    </Text>
    <div className="flex flex-wrap gap-2">
      <Badge variant="code" labelKey="env" labelValue="production" />
      <Badge variant="code" labelKey="region" labelValue="us-west-2" />
    </div>
  </div>
)

export const Default = () => (
  <div className="max-w-md p-8">
    <Card>
      <Metadata />
    </Card>
  </div>
)

export const OnBackdrop = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      The tint is see-through, so the card picks up whatever sits behind it
      while the blur keeps the text readable.
    </Text>
    <Backdrop>
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <Metadata />
        </Card>
        <Card blur="lg">
          <Metadata />
        </Card>
      </div>
    </Backdrop>
  </div>
)

export const Opacity = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      Three semantic opacity levels over the same backdrop.
    </Text>
    <Backdrop>
      <div className="grid gap-4 sm:grid-cols-3">
        {(['default', 'strong', 'solid'] as const).map((opacity) => (
          <Card key={opacity} opacity={opacity}>
            <Text variant="body" weight="medium">
              {opacity}
            </Text>
            <Text variant="caption" color="tertiary">
              Card background opacity
            </Text>
          </Card>
        ))}
      </div>
    </Backdrop>
  </div>
)

export const Blur = () => (
  <div className="flex flex-col gap-4 p-8">
    <Text variant="caption" color="tertiary">
      Four blur strengths over the same backdrop.
    </Text>
    <Backdrop>
      <div className="grid gap-4 sm:grid-cols-2">
        {(['none', 'sm', 'md', 'lg'] as const).map((blur) => (
          <Card key={blur} blur={blur}>
            <Text variant="body" weight="medium">
              blur {blur}
            </Text>
            <Text variant="caption" color="tertiary">
              Same tint, different backdrop filter.
            </Text>
          </Card>
        ))}
      </div>
    </Backdrop>
  </div>
)

export const Padding = () => (
  <div className="flex max-w-md flex-col gap-4 p-8">
    {(['none', 'sm', 'md', 'lg'] as const).map((padding) => (
      <Card key={padding} padding={padding}>
        <Text variant="caption" family="mono" color="tertiary">
          padding {padding}
        </Text>
      </Card>
    ))}
  </div>
)

export const Interactive = () => (
  <div className="grid max-w-2xl gap-4 p-8 sm:grid-cols-2">
    <Card interactive tabIndex={0}>
      <Metadata />
    </Card>
    <Card>
      <Metadata />
    </Card>
  </div>
)

export const FloatingShadow = () => (
  <div className="max-w-md bg-surface-02 p-12">
    <Card shadow="floating">
      <Metadata />
    </Card>
  </div>
)

export const WithActions = () => (
  <div className="max-w-md p-8">
    <Card>
      <div className="flex flex-col gap-4">
        <Metadata />
        <div className="flex justify-end gap-2">
          <Button size="sm" variant="secondary">
            View plan
          </Button>
          <Button size="sm">Deploy</Button>
        </div>
      </div>
    </Card>
  </div>
)

export const OverContent = () => (
  <div className="relative max-w-2xl p-8">
    <div aria-hidden className="flex flex-col gap-1">
      {Array.from({ length: 18 }).map((_, index) => (
        <Text key={index} variant="caption" family="mono" color="tertiary">
          {`2026-09-02T20:1${index % 10}:04Z  runner  reconciled component payments-api revision 41${index}`}
        </Text>
      ))}
    </div>
    <div className="absolute inset-x-16 top-24">
      <Card blur="lg">
        <Metadata />
      </Card>
    </div>
  </div>
)
