import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'
import { FleetFilterDropdown } from './FleetFilterDropdown'

export interface ILabelsFilterDropdown {
  options: string[]
  selected: Set<string>
  onChange: (selected: Set<string>) => void
}

export const LabelsFilterDropdown = ({
  options,
  selected,
  onChange,
}: ILabelsFilterDropdown) => {
  const toggle = (value: string) => {
    const next = new Set(selected)
    if (next.has(value)) next.delete(value)
    else next.add(value)
    onChange(next)
  }

  const only = (value: string) =>
    onChange(
      selected.size === 1 && selected.has(value) ? new Set() : new Set([value])
    )

  return (
    <FleetFilterDropdown
      id="fleet-filter-labels"
      label="Labels"
      count={selected.size}
      onReset={() => onChange(new Set())}
      menuClassName="!w-max max-w-80"
    >
      {options.map((value) => {
        const checked = selected.has(value)
        const isOnlySelected = selected.size === 1 && checked
        return (
          <CheckboxInputWithButton
            key={value}
            className="w-full min-w-0"
            name={`fleet-label-${value}`}
            value={value}
            checked={checked}
            onChange={() => toggle(value)}
            buttonProps={{
              className:
                '!p-1 flex items-center justify-between group w-full min-w-0 !rounded-sm',
              variant: 'ghost',
              type: 'button',
              value,
              title: value,
              onClick: () => only(value),
              children: (
                <>
                  <Text
                    variant="body"
                    weight="strong"
                    family="mono"
                    className="truncate"
                  >
                    {value}
                  </Text>
                  <Text
                    variant="subtext"
                    theme="neutral"
                    className="ml-2 shrink-0 opacity-0 group-hover:opacity-100"
                  >
                    {isOnlySelected ? 'Reset' : 'Only'}
                  </Text>
                </>
              ),
            }}
          />
        )
      })}
    </FleetFilterDropdown>
  )
}
