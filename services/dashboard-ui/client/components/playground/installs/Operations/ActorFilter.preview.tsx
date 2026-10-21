import { useMemo, useState } from 'react'
import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Toggle } from '@/components/common/form/Toggle'
import { cn } from '@/utils/classnames'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_LABELS,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'

// Org members only ever expose account id + email here — `app.Account.Name`
// is almost never populated outside manual/admin paths, and there's no
// avatar URL for anyone but the currently authenticated user.
type TMember = { accountId: string; email: string }

const ME: TMember = { accountId: 'acc-me', email: 'jane@example.com' }

const MEMBERS: TMember[] = [
  ME,
  { accountId: 'acc-1', email: 'joe@example.com' },
  { accountId: 'acc-2', email: 'priya@example.com' },
  { accountId: 'acc-3', email: 'marcus@example.com' },
  { accountId: 'acc-4', email: 'ana@example.com' },
  { accountId: 'acc-5', email: 'sam@example.com' },
  { accountId: 'acc-6', email: 'lena@example.com' },
  { accountId: 'acc-7', email: 'tom@example.com' },
  { accountId: 'acc-8', email: 'yuki@example.com' },
  { accountId: 'acc-9', email: 'omar@example.com' },
  { accountId: 'acc-10', email: 'grace@example.com' },
  { accountId: 'acc-11', email: 'diego@example.com' },
]

type TActivityType = 'action_run' | 'runbook_run'

type TMockActivity = {
  id: string
  type: TActivityType
  title: string
  status: string
  statusGroup: TWorkflowStatusOption
  hoursAgo: number
  actorId?: string
  cron?: boolean
}

const ACTIVITY: TMockActivity[] = [
  { id: 'a1', type: 'action_run', title: 'Restart workers', status: 'finished', statusGroup: 'succeeded', hoursAgo: 1, actorId: 'acc-1' },
  { id: 'a2', type: 'action_run', title: 'Healthcheck', status: 'finished', statusGroup: 'succeeded', hoursAgo: 2, cron: true },
  { id: 'a3', type: 'runbook_run', title: 'Rotate database credentials', status: 'in-progress', statusGroup: 'running', hoursAgo: 3, actorId: 'acc-me' },
  { id: 'a4', type: 'action_run', title: 'Flush cache', status: 'error', statusGroup: 'failed', hoursAgo: 5, actorId: 'acc-2' },
  { id: 'a5', type: 'action_run', title: 'Healthcheck', status: 'finished', statusGroup: 'succeeded', hoursAgo: 6, cron: true },
  { id: 'a6', type: 'runbook_run', title: 'Reconcile drift', status: 'finished', statusGroup: 'succeeded', hoursAgo: 9, actorId: 'acc-1' },
  { id: 'a7', type: 'action_run', title: 'Export audit log', status: 'queued', statusGroup: 'queued', hoursAgo: 12, actorId: 'acc-3' },
  { id: 'a8', type: 'action_run', title: 'Restart workers', status: 'finished', statusGroup: 'succeeded', hoursAgo: 20, actorId: 'acc-me' },
  { id: 'a9', type: 'runbook_run', title: 'Scale node pool', status: 'cancelled', statusGroup: 'cancelled', hoursAgo: 30, actorId: 'acc-4' },
  { id: 'a10', type: 'action_run', title: 'Flush cache', status: 'finished', statusGroup: 'succeeded', hoursAgo: 40, actorId: 'acc-1' },
]

const TYPE_LABELS: Record<TActivityType, string> = {
  action_run: 'Action run',
  runbook_run: 'Runbook run',
}

const TYPE_ICON: Record<TActivityType, TIconVariant> = {
  action_run: 'TerminalWindowIcon',
  runbook_run: 'BookIcon',
}

const STATUS_OPTIONS: TWorkflowStatusOption[] = [
  'running',
  'queued',
  'succeeded',
  'failed',
  'cancelled',
]

const memberById = (accountId?: string) =>
  MEMBERS.find((member) => member.accountId === accountId)

const matchesQuery = (member: TMember, query: string) =>
  !query.trim() || member.email.toLowerCase().includes(query.trim().toLowerCase())

interface IMemberRow {
  member: TMember
  isMe?: boolean
  selected: boolean
  onSelect: () => void
}

const MemberRow = ({ member, isMe, selected, onSelect }: IMemberRow) => (
  <button
    type="button"
    role="menuitemradio"
    aria-checked={selected}
    onClick={onSelect}
    className={cn(
      'flex w-full shrink-0 cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-left',
      'hover:bg-cool-grey-500/8 dark:hover:bg-cool-grey-500/8',
      'focus-visible:outline-1 focus-visible:outline-primary-600',
      selected && 'bg-primary-500/8 dark:bg-primary-500/12'
    )}
  >
    <Avatar name={member.email} size="xs" shape="circle" />
    <Text variant="subtext" weight="strong" className="min-w-0 flex-1 truncate">
      {member.email}
      {isMe ? (
        <Text as="span" variant="subtext" theme="neutral">
          {' '}
          (you)
        </Text>
      ) : null}
    </Text>
    {selected ? <Icon variant="CheckIcon" size={14} /> : null}
  </button>
)

interface ITriggeredByDropdown {
  selected?: string
  onChange: (accountId?: string) => void
  isOpen?: boolean
}

export const TriggeredByDropdown = ({
  selected,
  onChange,
  isOpen,
}: ITriggeredByDropdown) => {
  const [query, setQuery] = useState('')
  const selectedMember = memberById(selected)
  const showMe = matchesQuery(ME, query)
  const teammates = MEMBERS.filter(
    (member) => member.accountId !== ME.accountId && matchesQuery(member, query)
  )

  return (
    <Dropdown
      alignment="left"
      closeOnBlur={false}
      id="activity-filter-actor"
      isOpen={isOpen}
      buttonText={
        <>
          {selectedMember ? (
            <Avatar name={selectedMember.email} size="xs" shape="circle" />
          ) : (
            <Icon variant="UserIcon" size={14} />
          )}
          {selectedMember
            ? selectedMember.accountId === ME.accountId
              ? 'Triggered by me'
              : selectedMember.email
            : 'Triggered by'}
        </>
      }
    >
      <div className="flex w-80 flex-col gap-2 p-2">
        <SearchInput
          aria-label="Search teammates by email"
          placeholder="Search email"
          labelClassName="!w-full"
          className="!min-w-0 w-full"
          value={query}
          onChange={setQuery}
          onClear={() => setQuery('')}
        />
        <div className="flex max-h-72 flex-col gap-0.5 overflow-y-auto">
          {showMe ? (
            <>
              <MemberRow
                member={ME}
                isMe
                selected={selected === ME.accountId}
                onSelect={() => onChange(ME.accountId)}
              />
              {teammates.length ? <hr className="my-1" /> : null}
            </>
          ) : null}
          {teammates.map((member) => (
            <MemberRow
              key={member.accountId}
              member={member}
              selected={selected === member.accountId}
              onSelect={() => onChange(member.accountId)}
            />
          ))}
          {!showMe && !teammates.length ? (
            <Text variant="subtext" theme="neutral" className="px-2 py-3">
              No teammates match “{query}”
            </Text>
          ) : null}
        </div>
        <hr />
        <Button
          disabled={!selected}
          isMenuButton
          onClick={() => onChange(undefined)}
          type="button"
          variant="ghost"
        >
          Clear
          <Icon variant="XIcon" />
        </Button>
      </div>
    </Dropdown>
  )
}

const MockActivityCard = ({ activity }: { activity: TMockActivity }) => {
  const actor = memberById(activity.actorId)
  return (
    <Card className="!p-4 !gap-2 !shadow-none">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0">
          <span className="text-cool-grey-400 shrink-0">
            <Icon variant={TYPE_ICON[activity.type]} size={16} />
          </span>
          <Text variant="body" weight="strong" className="truncate">
            {activity.type === 'runbook_run' ? 'Runbook' : 'Action'}:{' '}
            {activity.title}
          </Text>
        </div>
        <Status status={activity.status} />
      </div>
      <div className="flex items-center gap-x-6 gap-y-2 flex-wrap pl-7">
        <Text variant="subtext" theme="neutral">
          {activity.cron ? 'system (cron)' : actor?.email}
        </Text>
        <Time
          time={new Date(Date.now() - activity.hoursAgo * 3600_000).toISOString()}
          format="relative"
          variant="subtext"
          theme="neutral"
        />
      </div>
    </Card>
  )
}

interface IActorFilterPreview {
  initialActor?: string
}

export const ActorFilterPreview = ({ initialActor }: IActorFilterPreview) => {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<Set<string>>(new Set())
  const [type, setType] = useState<Set<string>>(new Set())
  const [showCron, setShowCron] = useState(true)
  const [date, setDate] = useState<string>()
  const [actor, setActor] = useState<string | undefined>(initialActor)

  const visible = useMemo(
    () =>
      ACTIVITY.filter(
        (item) =>
          (!actor || item.actorId === actor) &&
          (showCron || !item.cron) &&
          (!status.size || status.has(item.statusGroup)) &&
          (!type.size || type.has(item.type)) &&
          (!search || item.title.toLowerCase().includes(search.toLowerCase()))
      ),
    [actor, showCron, status, type, search]
  )

  const hasFilters =
    !!search || !!actor || !!date || !showCron || status.size + type.size > 0

  const clear = () => {
    setSearch('')
    setStatus(new Set())
    setType(new Set())
    setShowCron(true)
    setDate(undefined)
    setActor(undefined)
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <SearchInput
          aria-label="Search activity"
          placeholder="Search activity"
          value={search}
          onChange={setSearch}
          onClear={() => setSearch('')}
        />
        <CheckboxFilterDropdown
          id="activity-filter-status"
          label="Status"
          options={STATUS_OPTIONS.map((value) => ({
            value,
            label: WORKFLOW_STATUS_LABELS[value],
          }))}
          selected={status}
          onChange={setStatus}
        />
        <CheckboxFilterDropdown
          id="activity-filter-type"
          label="Type"
          options={(Object.keys(TYPE_LABELS) as TActivityType[]).map(
            (value) => ({ value, label: TYPE_LABELS[value] })
          )}
          selected={type}
          onChange={setType}
        />
        <TriggeredByDropdown selected={actor} onChange={setActor} />
        <RadioFilterDropdown
          id="activity-filter-date"
          label="Date"
          options={Object.entries(WORKFLOW_DATE_LABELS).map(
            ([value, label]) => ({ value, label })
          )}
          selected={date}
          onChange={setDate}
        />
        <Toggle
          checked={showCron}
          onChange={setShowCron}
          label="Show cron runs"
          className="px-2"
        />
        {hasFilters ? (
          <Button variant="ghost" onClick={clear}>
            Clear filters
          </Button>
        ) : null}
      </div>

      {visible.length ? (
        <div className="flex flex-col gap-3">
          {visible.map((item) => (
            <MockActivityCard key={item.id} activity={item} />
          ))}
        </div>
      ) : (
        <EmptyState
          emptyTitle="No activity found"
          emptyMessage="No activity matches the current filters. Try adjusting or clearing them."
          className="my-12"
        />
      )}
    </div>
  )
}
