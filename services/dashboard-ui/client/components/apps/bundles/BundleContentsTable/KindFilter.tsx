import React from 'react'
import { useSearchParams } from 'react-router'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { CheckboxInputWithButton } from '@/components/common/form/CheckboxInput'

export interface IKindGroup {
  key: string
  label: string
  kinds: string[]
}

export const KIND_GROUPS: IKindGroup[] = [
  { key: 'components', label: 'Components', kinds: ['component', 'image'] },
  { key: 'sandbox', label: 'Sandbox', kinds: ['sandbox'] },
  { key: 'actions', label: 'Actions', kinds: ['action_step'] },
  { key: 'runner', label: 'Runner', kinds: ['runner_binary', 'runner_image'] },
  { key: 'stack', label: 'Stack assets', kinds: ['stack_asset'] },
]

export const useSelectedKindGroups = () => {
  const [searchParams] = useSearchParams()
  const param = searchParams.get('kinds')?.trim() ?? ''
  const keys = param
    ? param.split(',').filter((key) => KIND_GROUPS.some((g) => g.key === key))
    : []
  return keys.length > 0 ? keys : KIND_GROUPS.map((g) => g.key)
}

export const KindFilter = ({ groups }: { groups: IKindGroup[] }) => {
  const [, setSearchParams] = useSearchParams()
  const selected = useSelectedKindGroups()
  const visible = groups.length > 0 ? groups : KIND_GROUPS
  const allSelected = visible.every((g) => selected.includes(g.key))

  const setKindsInUrl = (keys: string[]) => {
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        if (keys.length > 0 && keys.length < KIND_GROUPS.length) {
          params.set('kinds', keys.join(','))
        } else {
          params.delete('kinds')
        }
        return params
      },
      { replace: true }
    )
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    const next = e.target.checked
      ? Array.from(new Set([...selected, value]))
      : selected.filter((key) => key !== value)
    setKindsInUrl(next)
  }

  const handleOnly = (e: React.MouseEvent<HTMLButtonElement>) => {
    const value = e.currentTarget.value
    setKindsInUrl(selected.length === 1 && selected[0] === value ? [] : [value])
  }

  const handleReset = () => setKindsInUrl([])

  return (
    <Dropdown
      alignment="right"
      className="!p-2"
      closeOnBlur={false}
      id="bundle-kind-filter"
      buttonClassName="!p-1"
      buttonText={
        <>
          <Icon variant="FunnelIcon" size="14" /> Kind
          {!allSelected && ` (${selected.length})`}
        </>
      }
    >
      <Menu className="min-w-44">
        {visible.map((group) => (
          <div className="flex items-center" key={group.key}>
            <CheckboxInputWithButton
              buttonProps={{
                className:
                  '!p-1 flex items-center justify-between group w-full',
                children: (
                  <>
                    <span className="text-xs font-semibold">{group.label}</span>
                    <span className="ml-2 text-xs opacity-0 group-hover:opacity-100 w-[40px]">
                      {selected.length === 1 && selected[0] === group.key
                        ? 'Reset'
                        : 'Only'}
                    </span>
                  </>
                ),
                type: 'button',
                variant: 'ghost',
                value: group.key,
                onClick: handleOnly,
              }}
              className="w-full"
              name={group.key}
              onChange={handleChange}
              checked={selected.includes(group.key)}
              value={group.key}
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
