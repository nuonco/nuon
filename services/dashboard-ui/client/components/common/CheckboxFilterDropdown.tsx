import type { ChangeEvent, MouseEvent } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'

export interface IFilterDropdownOption {
  label: string
  value: string
}

export interface ICheckboxFilterDropdown {
  id: string
  label: string
  options: IFilterDropdownOption[]
  selected: Set<string>
  onChange: (selected: Set<string>) => void
}

export const CheckboxFilterDropdown = ({
  id,
  label,
  options,
  selected,
  onChange,
}: ICheckboxFilterDropdown) => {
  const handleToggle = (event: ChangeEvent<HTMLInputElement>) => {
    const next = new Set(selected)
    if (event.target.checked) next.add(event.target.value)
    else next.delete(event.target.value)
    onChange(next)
  }

  const handleOnly = (event: MouseEvent<HTMLButtonElement>) => {
    const value = event.currentTarget.value
    onChange(
      selected.size === 1 && selected.has(value) ? new Set() : new Set([value])
    )
  }

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id={id}
      buttonText={
        <>
          <Icon variant="FunnelIcon" size={14} />
          {label}
          {selected.size ? ` (${selected.size})` : ''}
        </>
      }
    >
      <Menu className="min-w-56 max-h-80">
        <div className="-mx-2 -mt-2 flex min-h-0 flex-col gap-0.5 overflow-y-auto p-2">
          {options.map((option) => (
            <CheckboxInputWithButton
              key={option.value}
              buttonProps={{
                children: (
                  <>
                    <span className="font-semibold text-xs">
                      {option.label}
                    </span>
                    <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                      {selected.size === 1 && selected.has(option.value)
                        ? 'Reset'
                        : 'Only'}
                    </span>
                  </>
                ),
                onClick: handleOnly,
                type: 'button',
                value: option.value,
              }}
              checked={selected.has(option.value)}
              className="w-full shrink-0"
              name={`${id}-${option.value}`}
              onChange={handleToggle}
              value={option.value}
            />
          ))}
        </div>
        <hr />
        <Button
          className="shrink-0"
          disabled={selected.size === 0}
          isMenuButton
          onClick={() => onChange(new Set())}
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
