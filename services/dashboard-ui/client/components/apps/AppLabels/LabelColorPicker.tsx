import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { ColorDot } from './ColorDot'

const SWATCH_COLORS = [
  '#2563eb', '#dc2626', '#16a34a', '#9333ea', '#ca8a04', '#0891b2',
  '#e11d48', '#4f46e5', '#059669', '#c026d3', '#d97706', '#0284c7',
  '#7c3aed', '#15803d', '#a21caf', '#b45309', '#6366f1', '#ef4444',
  '#22c55e', '#a855f7', '#eab308', '#06b6d4', '#f43f5e', '#818cf8',
]

interface ILabelColorPicker {
  id: string
  value?: string
  defaultColor?: string
  isOverride?: boolean
  disabled?: boolean
  ariaLabel?: string
  onSelect: (color: string) => void
  onReset?: () => void
}

export const LabelColorPicker = ({
  id,
  value,
  defaultColor,
  isOverride,
  disabled,
  ariaLabel = 'Change color',
  onSelect,
  onReset,
}: ILabelColorPicker) => (
  <Dropdown
    id={id}
    variant="ghost"
    buttonText={
      <>
        <ColorDot color={value} title={value} />
        <span className="sr-only">{ariaLabel}</span>
      </>
    }
    disabled={disabled}
    position="below"
    alignment="right"
    dropdownClassName="p-3"
  >
    <div className="flex flex-col gap-2 w-56">
      <Text variant="label" weight="strong">Pick a color</Text>
      <div className="flex flex-wrap gap-1.5">
        {SWATCH_COLORS.map((color) => {
          const active = value?.toLowerCase() === color.toLowerCase()
          return (
            <button
              key={color}
              type="button"
              aria-label={`Use color ${color}`}
              aria-pressed={active}
              title={color}
              onClick={() => onSelect(color)}
              className={cn(
                'w-5 h-5 rounded-sm border-2 cursor-pointer transition-transform hover:scale-110',
                active ? 'border-dark-grey-950 dark:border-white scale-110' : 'border-transparent'
              )}
              style={{ backgroundColor: color }}
            />
          )
        })}
        <label className="flex items-center cursor-pointer" title="Custom color">
          <span className="sr-only">Custom color</span>
          <input
            type="color"
            className="w-5 h-5 rounded-sm cursor-pointer bg-transparent"
            defaultValue={value ?? defaultColor}
            onChange={(e) => onSelect(e.target.value)}
          />
        </label>
      </div>
      {isOverride && onReset && (
        <Button variant="ghost" onClick={onReset} className="self-start">
          <Icon variant="ArrowCounterClockwiseIcon" size={14} />
          Reset to default
        </Button>
      )}
    </div>
  </Dropdown>
)
