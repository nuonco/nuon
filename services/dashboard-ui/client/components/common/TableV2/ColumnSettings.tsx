import type { ChangeEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'

export interface IColumnSettingsOption {
  id: string
  label: string
  locked: boolean
  visible: boolean
}

export interface IColumnSettings {
  canReset: boolean
  id: string
  onReset: () => void
  onToggle: (id: string, visible: boolean) => void
  options: IColumnSettingsOption[]
}

export const ColumnSettings = ({
  canReset,
  id,
  onReset,
  onToggle,
  options,
}: IColumnSettings) => {
  const handleChange = (event: ChangeEvent<HTMLInputElement>) =>
    onToggle(event.target.value, event.target.checked)

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id={id}
      hideIcon
      size="sm"
      aria-label="Table settings"
      tooltipProps={{
        className: 'flex',
        position: 'bottom',
        tipContent: 'Edit columns',
      }}
      buttonText={<Icon variant="SlidersHorizontalIcon" size={16} />}
    >
      <Menu className="!w-max min-w-56 max-w-80">
        <Text>Show columns</Text>
        <div className="-mx-2 flex max-h-80 min-h-0 flex-col gap-0.5 overflow-y-auto px-2">
          {options.map((option) => (
            <CheckboxInput
              key={option.id}
              checked={option.visible}
              disabled={option.locked}
              name={`${id}-${option.id}`}
              onChange={handleChange}
              value={option.id}
              labelProps={{
                className: option.locked ? '!cursor-not-allowed opacity-60' : '',
                labelText: option.label,
                labelTextProps: { variant: 'subtext' },
              }}
            />
          ))}
        </div>
        <hr />
        <Button
          disabled={!canReset}
          isMenuButton
          onClick={onReset}
          type="button"
          variant="ghost"
        >
          Reset to default
          <Icon variant="ArrowCounterClockwiseIcon" />
        </Button>
      </Menu>
    </Dropdown>
  )
}
