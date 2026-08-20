import React from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'

const NONE_TOKEN = '__none__'
const NONE_LABEL = 'No branch'

interface IInstallBranchFilter {
  queryKey: string[]
  queryFn: () => Promise<string[]>
}

export const InstallBranchFilter = ({
  queryKey,
  queryFn,
}: IInstallBranchFilter) => {
  const [searchParams, setSearchParams] = useSearchParams()

  const { data: branchNames } = useQuery({
    placeholderData: keepPreviousData,
    queryKey,
    queryFn,
  })

  const options = [...(branchNames ?? []), NONE_TOKEN]

  const param = searchParams.get('branches')
  const allSelected = !param
  const selected = allSelected
    ? options
    : param
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean)

  const setBranchesInUrl = (values: string[]) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        if (values.length === 0 || values.length === options.length) {
          params.delete('branches')
        } else {
          params.set('branches', values.join(','))
        }
        params.delete('offset')
        return params
      },
      { replace: true }
    )
  }

  const handleToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (e.target.checked) {
      setBranchesInUrl(Array.from(new Set([...selected, value])))
    } else {
      setBranchesInUrl(selected.filter((v) => v !== value))
    }
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    setBranchesInUrl([e.currentTarget.value])
  }

  const handleShowAll = () => setBranchesInUrl(options)

  if ((branchNames ?? []).length === 0) return null

  const labelFor = (value: string) =>
    value === NONE_TOKEN ? NONE_LABEL : value

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id="branches-filter"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" />
          Branch{!allSelected ? ` (${selected.length})` : ''}
        </>
      }
    >
      <Menu className="min-w-64 max-h-80 overflow-y-auto">
        <Text variant="label" theme="neutral" className="px-1">
          Filter by branch
        </Text>

        {options.map((value) => {
          const isOnlySelected = selected.length === 1 && selected[0] === value
          return (
            <div className="flex items-center space-x-2" key={value}>
              <CheckboxInputWithButton
                buttonProps={{
                  className:
                    '!p-1 flex items-center justify-between group w-full',
                  children: (
                    <>
                      <span className="font-semibold text-xs">
                        {labelFor(value)}
                      </span>
                      <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                        {isOnlySelected ? 'Reset' : 'Only'}
                      </span>
                    </>
                  ),
                  type: 'button',
                  variant: 'ghost',
                  value,
                  onClick: isOnlySelected ? handleShowAll : handleOnly,
                }}
                className="w-full"
                name={value}
                onChange={handleToggle}
                checked={selected.includes(value)}
                value={value}
              />
            </div>
          )
        })}

        <hr />

        <Button
          className="w-full !p-1 shrink-0"
          type="button"
          onClick={handleShowAll}
          size="sm"
          variant="ghost"
        >
          Reset
        </Button>
      </Menu>
    </Dropdown>
  )
}
