import { Icon, type TIconVariant } from './Icon'
import { Text } from './Text'
import { ComponentDocs } from '../__stories__/ComponentDocs'

export default {
  title: 'lite/atoms/Icon',
}

const VARIANTS: TIconVariant[] = [
  'ArrowClockwiseIcon',
  'ArrowSquareOutIcon',
  'CaretDownIcon',
  'CaretRightIcon',
  'CheckIcon',
  'CircleHalfIcon',
  'DesktopIcon',
  'MoonIcon',
  'PlusIcon',
  'SunIcon',
  'TrashIcon',
  'XIcon',
]

export const Overview = () => (
  <ComponentDocs
    name="Icon"
    tier="atom"
    summary="Phosphor icons behind a static map, so only the icons lite uses are bundled."
    use={['Any icon in lite. Add missing ones to the ICONS map.']}
    avoid={[
      'Importing from @phosphor-icons/react at a call site — it defeats the map and the shared defaults. A lint rule blocks it.',
      "The dashboard's common/Icon, which pulls in react-icons for three logos and ~200 entries lite does not use.",
    ]}
    rules={[
      'Icons are currentColor — they take their colour from the text around them, never a colour prop.',
      'The map is static so the bundler can tree-shake. Adding an icon means an import plus a map entry.',
      'Phosphor glyphs sit inside a 256 viewBox with transparent padding — about 12.5% a side. Cancel it with a negative margin when an icon sits at the edge of a control.',
    ]}
    props={[
      {
        name: 'variant',
        type: 'TIconVariant',
        description: 'Key from the ICONS map.',
      },
      {
        name: 'size',
        type: 'number | string',
        default: '16',
        description: 'Use "1em" to track surrounding font size.',
      },
      {
        name: 'weight',
        type: "'thin' | 'light' | 'regular' | 'bold' | 'fill' | 'duotone'",
        default: "'regular'",
        description: 'Phosphor stroke weight.',
      },
    ]}
  />
)

export const All = () => (
  <div className="grid grid-cols-4 gap-4 p-8">
    {VARIANTS.map((variant) => (
      <div key={variant} className="flex items-center gap-2">
        <Icon variant={variant} />
        <Text variant="caption" color="tertiary" family="mono">
          {variant.replace('Icon', '')}
        </Text>
      </div>
    ))}
  </div>
)

export const Sizes = () => (
  <div className="flex items-center gap-6 p-8">
    {[12, 16, 20, 24, 32].map((size) => (
      <div key={size} className="flex flex-col items-center gap-2">
        <Icon variant="SunIcon" size={size} />
        <Text variant="label" color="tertiary" family="mono">
          {size}
        </Text>
      </div>
    ))}
  </div>
)

export const Weights = () => (
  <div className="flex items-center gap-6 p-8">
    {(['thin', 'light', 'regular', 'bold', 'fill'] as const).map((weight) => (
      <div key={weight} className="flex flex-col items-center gap-2">
        <Icon variant="CheckIcon" size={24} weight={weight} />
        <Text variant="label" color="tertiary" family="mono">
          {weight}
        </Text>
      </div>
    ))}
  </div>
)

export const InheritsColor = () => (
  <div className="flex flex-col gap-3 p-8">
    <Text color="primary" className="flex items-center gap-2">
      <Icon variant="CheckIcon" /> primary
    </Text>
    <Text color="tertiary" className="flex items-center gap-2">
      <Icon variant="CheckIcon" /> tertiary
    </Text>
    <Text color="accent" className="flex items-center gap-2">
      <Icon variant="CheckIcon" /> accent
    </Text>
    <Text variant="caption" color="tertiary">
      Icons are currentColor and inherit font-size when size is set to “1em”.
    </Text>
  </div>
)
