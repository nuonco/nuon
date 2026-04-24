import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Button } from './Button'
import { Dropdown } from './Dropdown'
import { Icon } from './Icon'
import { Menu } from './Menu'
import { Text } from './Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'

interface ILabelFilterDropdown {
  queryKey: string[]
  queryFn: () => Promise<Record<string, string[]>>
}

export const LabelFilterDropdown = ({
  queryKey,
  queryFn,
}: ILabelFilterDropdown) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const { data: labelMap } = useQuery({
    queryKey,
    queryFn,
  })

  const currentLabels = searchParams.get('labels') || ''
  const selectedLabels = currentLabels
    ? currentLabels.split(',').map((l) => l.trim())
    : []

  const setLabelsInUrl = (labels: string[]) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        if (labels.length > 0) {
          params.set('labels', labels.join(','))
        } else {
          params.delete('labels')
        }
        params.delete('offset')
        return params
      },
      { replace: true },
    )
  }

  const handleToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (e.target.checked) {
      setLabelsInUrl(Array.from(new Set([...selectedLabels, value])))
    } else {
      setLabelsInUrl(selectedLabels.filter((l) => l !== value))
    }
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    const value = e.currentTarget.value
    setLabelsInUrl([value])
  }

  const handleReset = () => setLabelsInUrl([])

  if (!labelMap || Object.keys(labelMap).length === 0) return null

  const allOptions: string[] = []
  for (const key of Object.keys(labelMap).sort()) {
    for (const val of labelMap[key]) {
      allOptions.push(`${key}:${val}`)
    }
  }

  return (
    <Dropdown
      alignment="right"
      className="!p-2"
      closeOnBlur={false}
      id="labels-filter"
      buttonClassName="!p-1"
      buttonText={
        <>
          <Icon variant="TagIcon" size="14" />
          Labels{selectedLabels.length > 0 ? ` (${selectedLabels.length})` : ''}
        </>
      }
    >
      <Menu className="min-w-64 max-h-80 overflow-y-auto">
        <Text variant="label" theme="neutral" className="px-1">
          Filter by label
        </Text>

        {allOptions.map((label) => (
          <div className="flex items-center space-x-2" key={label}>
            <CheckboxInputWithButton
              buttonProps={{
                className: '!p-1 flex items-center justify-between group w-full',
                children: (
                  <>
                    <span className="font-semibold text-xs font-mono">
                      {label}
                    </span>
                    <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                      {selectedLabels.length === 1 &&
                      selectedLabels[0] === label
                        ? 'Reset'
                        : 'Only'}
                    </span>
                  </>
                ),
                type: 'button',
                variant: 'ghost',
                value: label,
                onClick:
                  selectedLabels.length === 1 && selectedLabels[0] === label
                    ? handleReset
                    : handleOnly,
              }}
              className="w-full"
              name={label}
              onChange={handleToggle}
              checked={selectedLabels.includes(label)}
              value={label}
            />
          </div>
        ))}

        <hr />

        <Button
          className="w-full !p-1 shrink-0"
          type="button"
          onClick={handleReset}
          size="sm"
          variant="ghost"
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
