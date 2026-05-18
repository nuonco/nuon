import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Text } from '@/components/common/Text'
import type { TInstall } from '@/types'

interface IAddInstallPicker {
  groupId: string
  unassignedInstalls: TInstall[]
  disabled?: boolean
  onAdd: (installIds: string[]) => void
}

const SEARCH_THRESHOLD = 5

export const AddInstallPicker = ({
  groupId,
  unassignedInstalls,
  disabled,
  onAdd,
}: IAddInstallPicker) => {
  const [picked, setPicked] = useState<Set<string>>(new Set())
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => {
    if (!query) return unassignedInstalls
    const q = query.toLowerCase()
    return unassignedInstalls.filter((i) => i.name.toLowerCase().includes(q))
  }, [unassignedInstalls, query])

  const toggle = (id: string) => {
    setPicked((curr) => {
      const next = new Set(curr)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const commit = () => {
    if (picked.size === 0) return
    onAdd(Array.from(picked))
    setPicked(new Set())
    setQuery('')
  }

  const isEmpty = unassignedInstalls.length === 0

  return (
    <Dropdown
      id={`add-install-${groupId}`}
      variant="secondary"
      alignment="left"
      closeOnBlur={false}
      hideIcon
      disabled={disabled || isEmpty}
      buttonClassName="self-start"
      buttonText={
        <span className="inline-flex items-center gap-1.5">
          <Icon variant="PlusIcon" size={14} />
          {isEmpty ? 'All installs assigned' : 'Add install'}
        </span>
      }
    >
      <div className="flex flex-col w-[320px]">
        {unassignedInstalls.length > SEARCH_THRESHOLD && (
          <div className="p-3 border-b border-cool-grey-200 dark:border-dark-grey-700">
            <Input
              id={`pick-search-${groupId}`}
              type="text"
              size="sm"
              placeholder="Search installs..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              aria-label="Search installs"
            />
          </div>
        )}

        <div className="flex flex-col max-h-[260px] overflow-y-auto p-2 gap-0.5">
          {filtered.length === 0 ? (
            <div className="px-2 py-4 text-center">
              <Text variant="subtext" theme="neutral">
                {query
                  ? 'No installs match your search.'
                  : 'No unassigned installs.'}
              </Text>
            </div>
          ) : (
            filtered.map((install) => (
              <CheckboxInput
                key={install.id}
                id={`pick-${groupId}-${install.id}`}
                checked={picked.has(install.id)}
                onChange={() => toggle(install.id)}
                labelProps={{
                  labelText: install.name,
                }}
              />
            ))
          )}
        </div>

        <div className="flex items-center justify-end gap-2 p-3 border-t border-cool-grey-200 dark:border-dark-grey-700">
          <Text variant="subtext" theme="neutral" className="mr-auto">
            {picked.size > 0 ? `${picked.size} selected` : ''}
          </Text>
          <Button
            variant="primary"
            size="sm"
            onClick={commit}
            disabled={picked.size === 0}
          >
            {picked.size > 0 ? `Add ${picked.size}` : 'Add'}
          </Button>
        </div>
      </div>
    </Dropdown>
  )
}
