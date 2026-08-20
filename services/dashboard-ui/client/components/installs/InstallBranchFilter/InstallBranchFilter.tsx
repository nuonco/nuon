import React from 'react'
import { useSearchParams } from 'react-router'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'

type TBranchStatus = 'assigned' | 'none'

const FILTER_OPTIONS: Array<{ value: TBranchStatus; label: string }> = [
  { value: 'assigned', label: 'Assigned to a branch' },
  { value: 'none', label: 'No branch' },
]

const ALL_VALUES = FILTER_OPTIONS.map((o) => o.value)

export const InstallBranchFilter = () => {
  const [searchParams, setSearchParams] = useSearchParams()

  const param = searchParams.get('branch_status')
  const allSelected = !param
  const selected: TBranchStatus[] = allSelected
    ? ALL_VALUES
    : param
        .split(',')
        .map((v) => v.trim())
        .filter((v): v is TBranchStatus =>
          ALL_VALUES.includes(v as TBranchStatus)
        )

  const setStatusInUrl = (values: TBranchStatus[]) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        if (values.length === 0 || values.length === ALL_VALUES.length) {
          params.delete('branch_status')
        } else {
          params.set('branch_status', values.join(','))
        }
        params.delete('offset')
        return params
      },
      { replace: true }
    )
  }

  const handleToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value as TBranchStatus
    if (e.target.checked) {
      setStatusInUrl(Array.from(new Set([...selected, value])))
    } else {
      setStatusInUrl(selected.filter((v) => v !== value))
    }
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    setStatusInUrl([e.currentTarget.value as TBranchStatus])
  }

  const handleShowAll = () => setStatusInUrl(ALL_VALUES)

  return (
    <Dropdown
      alignment="right"
      closeOnBlur={false}
      id="branch-status-filter"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" />
          Branch{!allSelected ? ` (${selected.length})` : ''}
        </>
      }
    >
      <Menu className="min-w-64">
        {FILTER_OPTIONS.map((opt) => {
          const isOnlySelected =
            selected.length === 1 && selected[0] === opt.value
          return (
            <div className="flex items-center space-x-2" key={opt.value}>
              <CheckboxInputWithButton
                buttonProps={{
                  className:
                    '!p-1 flex items-center justify-between group w-full',
                  children: (
                    <>
                      <span className="font-semibold text-xs">{opt.label}</span>
                      <span className="ml-2 text-xs opacity-0 group-hover:opacity-100">
                        {isOnlySelected ? 'Reset' : 'Only'}
                      </span>
                    </>
                  ),
                  type: 'button',
                  variant: 'ghost',
                  value: opt.value,
                  onClick: isOnlySelected ? handleShowAll : handleOnly,
                }}
                className="w-full"
                name={opt.value}
                onChange={handleToggle}
                checked={selected.includes(opt.value)}
                value={opt.value}
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
