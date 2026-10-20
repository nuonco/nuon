import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IActorFilterOption {
  accountId: string
  email: string
}

interface IActorRow {
  member: IActorFilterOption
  isMe?: boolean
  selected: boolean
  onSelect: () => void
}

const ActorRow = ({ member, isMe, selected, onSelect }: IActorRow) => (
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

export interface IActorFilter {
  id: string
  currentUser?: IActorFilterOption
  members: IActorFilterOption[]
  isLoading?: boolean
  query: string
  onQueryChange: (value: string) => void
  selected?: string
  onChange: (accountId?: string) => void
}

export const ActorFilter = ({
  id,
  currentUser,
  members,
  isLoading,
  query,
  onQueryChange,
  selected,
  onChange,
}: IActorFilter) => {
  const selectedMember =
    currentUser?.accountId === selected
      ? currentUser
      : members.find((member) => member.accountId === selected)
  const showMe =
    !!currentUser &&
    (!query.trim() ||
      currentUser.email.toLowerCase().includes(query.trim().toLowerCase()))

  return (
    <Dropdown
      alignment="left"
      closeOnBlur={false}
      id={id}
      buttonText={
        <>
          {selectedMember ? (
            <Avatar name={selectedMember.email} size="xs" shape="circle" />
          ) : (
            <Icon variant="UserIcon" size={14} />
          )}
          {selectedMember
            ? selectedMember.accountId === currentUser?.accountId
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
          onChange={onQueryChange}
          onClear={() => onQueryChange('')}
        />
        <div className="flex max-h-72 flex-col gap-0.5 overflow-y-auto">
          {showMe ? (
            <>
              <ActorRow
                member={currentUser}
                isMe
                selected={selected === currentUser.accountId}
                onSelect={() => onChange(currentUser.accountId)}
              />
              {members.length ? <hr className="my-1" /> : null}
            </>
          ) : null}
          {members.map((member) => (
            <ActorRow
              key={member.accountId}
              member={member}
              selected={selected === member.accountId}
              onSelect={() => onChange(member.accountId)}
            />
          ))}
          {isLoading ? (
            <Text variant="subtext" theme="neutral" className="px-2 py-3">
              Searching…
            </Text>
          ) : !showMe && !members.length ? (
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
