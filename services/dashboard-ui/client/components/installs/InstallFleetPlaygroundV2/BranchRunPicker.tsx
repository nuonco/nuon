import { useState, type ReactNode } from 'react'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RadioInput } from '@/components/common/form/RadioInput'

const SEARCH_THRESHOLD = 5

export type TFleetRunOption = {
  value: string
  title: string
  sha?: string
  createdAt?: string
  count: number
}

export interface IBranchRunPicker {
  id: string
  runs: TFleetRunOption[]
  selected?: string
  onChange: (value?: string) => void
}

const plural = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

export const BranchRunPicker = ({
  id,
  runs,
  selected,
  onChange,
}: IBranchRunPicker) => {
  const [search, setSearch] = useState('')
  const latest = runs[0]
  const picked = runs.find((run) => run.value === selected)
  const buttonLabel = picked
    ? picked.title
    : `Latest run${latest ? ` · ${latest.title}` : ''}`
  const needle = search.trim().toLowerCase()
  const shown = needle
    ? runs.filter(
        (run) =>
          run.title.toLowerCase().includes(needle) ||
          run.sha?.toLowerCase().includes(needle)
      )
    : runs

  return (
    <Dropdown
      id={id}
      variant="ghost"
      buttonText={
        <>
          <Icon variant="GitCommitIcon" size={14} />
          <span className="min-w-0 max-w-64 truncate" title={buttonLabel}>
            {buttonLabel}
          </span>
        </>
      }
      closeOnBlur={false}
    >
      <Menu className="!w-80 max-h-[28rem]">
        {runs.length > SEARCH_THRESHOLD ? (
          <SearchInput
            labelClassName="w-full"
            className="w-full"
            placeholder="Search runs"
            value={search}
            onChange={setSearch}
          />
        ) : null}
        <div className="-mx-2 flex min-h-0 flex-col gap-0.5 overflow-y-auto overflow-x-hidden px-2">
          {!needle ? (
            <RadioInput
              name={id}
              checked={!picked}
              onChange={() => onChange(undefined)}
              labelProps={{
                labelText: (
                  <RunLabel
                    title="Latest run"
                    meta={latest ? `Follows ${latest.title}` : 'No runs yet'}
                  />
                ),
              }}
            />
          ) : null}
          {shown.map((run) => (
            <RadioInput
              key={run.value}
              name={id}
              checked={picked?.value === run.value}
              onChange={() => onChange(run.value)}
              labelProps={{
                title: run.title,
                labelText: (
                  <RunLabel
                    title={run.title}
                    meta={
                      <>
                        {run.sha ? (
                          <span className="font-mono">
                            {run.sha.slice(0, 7)}
                          </span>
                        ) : null}
                        {run.createdAt ? (
                          <>
                            {' · '}
                            <Time
                              time={run.createdAt}
                              format="relative"
                              variant="subtext"
                              theme="neutral"
                            />
                          </>
                        ) : null}
                        {` · ${plural(run.count, 'install')}`}
                      </>
                    }
                  />
                ),
              }}
            />
          ))}
          {needle && !shown.length ? (
            <Text variant="subtext" theme="neutral" className="px-2 py-1">
              No runs match.
            </Text>
          ) : null}
        </div>
      </Menu>
    </Dropdown>
  )
}

const RunLabel = ({ title, meta }: { title: string; meta: ReactNode }) => (
  <span className="flex min-w-0 flex-col gap-1">
    <Text variant="body" weight="strong" className="truncate" nowrap>
      {title}
    </Text>
    <Text variant="subtext" theme="neutral" nowrap>
      {meta}
    </Text>
  </span>
)
