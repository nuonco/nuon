import { Button } from '@/components/common/Button'
import type { IFilterDropdownOption } from '@/components/common/CheckboxFilterDropdown'
import { Dropdown } from '@/components/common/Dropdown'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

export interface IRadioFilterDropdown {
  id: string
  label: string
  options: IFilterDropdownOption[]
  selected?: string
  onChange: (selected?: string) => void
}

export const RadioFilterDropdown = ({
  id,
  label,
  options,
  selected,
  onChange,
}: IRadioFilterDropdown) => {
  const selectedLabel = options.find(
    (option) => option.value === selected
  )?.label

  return (
    <Dropdown
      alignment="right"
      id={id}
      buttonText={
        <>
          <Icon variant="FunnelIcon" size={14} />
          {selectedLabel ?? label}
        </>
      }
    >
      <Menu className="min-w-56 max-h-80">
        <div className="-mx-2 -mt-2 flex min-h-0 flex-col gap-0.5 overflow-y-auto p-2">
          {options.map((option) => (
            <RadioInput
              key={option.value}
              checked={selected === option.value}
              labelProps={{ className: 'shrink-0', labelText: option.label }}
              name={id}
              onChange={(event) => onChange(event.target.value)}
              value={option.value}
            />
          ))}
        </div>
        <hr />
        <Button
          className="shrink-0"
          disabled={!selected}
          isMenuButton
          onClick={() => onChange(undefined)}
          type="button"
          variant="ghost"
        >
          Clear
          <Icon variant="XIcon" />
        </Button>
      </Menu>
    </Dropdown>
  )
}
