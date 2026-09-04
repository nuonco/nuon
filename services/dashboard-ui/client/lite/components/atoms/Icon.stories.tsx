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
      'Do not import from @phosphor-icons/react at a call site. A lint rule blocks it.',
      'Do not pass a colour. Set the colour on the surrounding text instead.',
    ]}
    rules={[
      'Icons use the current text colour and take their colour from the text around them.',
      'Adding an icon takes an import plus a map entry. The map is static so the bundler can tree-shake it.',
      'Set the size to one em to track the surrounding font size.',
      'Phosphor glyphs carry transparent padding inside their box, so pull an icon toward the edge with a negative margin when it sits at the edge of a control.',
    ]}
    props={[
      { name: 'variant', type: 'TIconVariant', description: 'Key from the ICONS map.' },
      { name: 'size', type: 'number | string', default: '16', description: 'Pixel size, or "1em" to follow the text.' },
      { name: 'weight', type: "'thin' | 'light' | 'regular' | 'bold' | 'fill' | 'duotone'", default: "'regular'", description: 'Phosphor stroke weight.' },
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
