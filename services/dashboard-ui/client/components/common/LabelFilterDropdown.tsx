import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from './Badge'
import { Button } from './Button'
import { Dropdown } from './Dropdown'
import { Icon } from './Icon'
import { Menu } from './Menu'
import { Text } from './Text'

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

  const handleOnly = (label: string) => {
    if (selectedLabels.length === 1 && selectedLabels[0] === label) {
      setLabelsInUrl([])
    } else {
      setLabelsInUrl([label])
    }
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
          Labels
          {selectedLabels.length > 0 ? (
            <Badge size="sm" theme="brand">
              {selectedLabels.length}
            </Badge>
          ) : null}
        </>
      }
    >
      <Menu className="min-w-64 max-h-80 overflow-y-auto">
        <Text variant="label" theme="neutral" className="px-1">
          Filter by label
        </Text>

        {allOptions.map((label) => {
          const isSelected =
            selectedLabels.length === 1 && selectedLabels[0] === label

          return (
            <Button
              key={label}
              className="!p-1 flex items-center justify-between w-full group"
              type="button"
              variant="ghost"
              onClick={() => handleOnly(label)}
            >
              <Badge variant="code" size="sm" theme={isSelected ? 'brand' : 'neutral'}>
                {label}
              </Badge>
              <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                {isSelected ? 'Reset' : 'Only'}
              </span>
            </Button>
          )
        })}

        {selectedLabels.length > 0 ? (
          <>
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
          </>
        ) : null}
      </Menu>
    </Dropdown>
  )
}
