import React from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'

export type TRepoFilterType = 'public' | 'private' | 'fork'

export const REPO_FILTER_OPTIONS: TRepoFilterType[] = ['public', 'private', 'fork']

const FILTER_LABELS: Record<TRepoFilterType, React.ReactNode> = {
  public: <><Icon variant="GlobeIcon" size={12} /> public</>,
  private: <><Icon variant="LockIcon" size={12} /> private</>,
  fork: <><Icon variant="GitForkIcon" size={12} /> fork</>,
}

interface IRepoTypeFilter {
  selected: TRepoFilterType[]
  onChange: (selected: TRepoFilterType[]) => void
}

export const RepoTypeFilter = ({ selected, onChange }: IRepoTypeFilter) => {
  const allSelected = selected.length === REPO_FILTER_OPTIONS.length

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value as TRepoFilterType
    const next = e.target.checked
      ? Array.from(new Set([...selected, value]))
      : selected.filter((t) => t !== value)
    onChange(next.length === 0 ? REPO_FILTER_OPTIONS : next)
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    const value = e.currentTarget.value as TRepoFilterType
    onChange(selected.length === 1 && selected[0] === value ? REPO_FILTER_OPTIONS : [value])
  }

  const handleReset = () => onChange(REPO_FILTER_OPTIONS)

  return (
    <Dropdown
      alignment="right"
      className="!p-2"
      closeOnBlur={false}
      id="repo-type-filter"
      buttonClassName="!p-1"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" /> Filter{!allSelected && ` (${selected.length})`}
        </>
      }
    >
      <Menu className="min-w-40">
        {REPO_FILTER_OPTIONS.map((opt) => (
          <div className="flex items-center" key={opt}>
            <CheckboxInputWithButton
              buttonProps={{
                className: '!p-1 flex items-center justify-between group w-full',
                children: (
                  <>
                    <span className="flex items-center gap-1 text-xs font-semibold">
                      {FILTER_LABELS[opt]}
                    </span>
                    <span className="ml-2 text-xs opacity-0 group-hover:opacity-100 w-[40px]">
                      {selected.length === 1 && selected[0] === opt ? 'Reset' : 'Only'}
                    </span>
                  </>
                ),
                type: 'button',
                variant: 'ghost',
                value: opt,
                onClick: handleOnly,
              }}
              className="w-full"
              name={opt}
              onChange={handleChange}
              checked={selected.includes(opt)}
              value={opt}
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
