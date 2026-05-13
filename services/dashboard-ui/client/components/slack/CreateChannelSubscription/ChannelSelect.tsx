import { useEffect, useMemo, useRef, useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import type { TSlackChannel } from '@/types'
import { cn } from '@/utils/classnames'

const SCROLL_FETCH_THRESHOLD = 64

export interface IChannelSelect {
  id?: string
  channels: TSlackChannel[]
  value: string
  onChange: (channelId: string, channelName: string) => void
  searchQuery: string
  onSearchChange: (q: string) => void
  onLoadMore: () => void
  hasMore: boolean
  isLoadingFirstPage: boolean
  isFetchingNextPage: boolean
  disabled?: boolean
  placeholder?: string
}

export const ChannelSelect = ({
  id,
  channels,
  value,
  onChange,
  searchQuery,
  onSearchChange,
  onLoadMore,
  hasMore,
  isLoadingFirstPage,
  isFetchingNextPage,
  disabled,
  placeholder,
}: IChannelSelect) => {
  const [isOpen, setIsOpen] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  const selected = channels.find((c) => c.id === value)

  const filtered = useMemo(() => {
    if (!searchQuery.trim()) return channels
    const q = searchQuery.trim().toLowerCase()
    return channels.filter((c) => (c.name ?? '').toLowerCase().includes(q))
  }, [channels, searchQuery])

  useEffect(() => {
    if (!isOpen) return
    const handleClickOutside = (e: MouseEvent) => {
      if (!wrapperRef.current?.contains(e.target as Node)) setIsOpen(false)
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  const handleScroll = () => {
    const el = listRef.current
    if (!el) return
    if (!hasMore || isFetchingNextPage) return
    const remaining = el.scrollHeight - el.scrollTop - el.clientHeight
    if (remaining < SCROLL_FETCH_THRESHOLD) onLoadMore()
  }

  const buttonLabel = selected?.name
    ? `#${selected.name}`
    : isLoadingFirstPage
      ? 'Loading channels…'
      : (placeholder ?? 'Select a channel')

  return (
    <div className="relative" ref={wrapperRef}>
      <button
        id={id}
        type="button"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        disabled={disabled}
        onClick={() => !disabled && setIsOpen((v) => !v)}
        className={cn(
          'flex items-center justify-between w-full border border-solid rounded-md px-3 py-2 text-sm transition-all duration-300 font-sans',
          'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]',
          'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:!border-primary-500',
          {
            '!bg-cool-grey-200 text-cool-grey-500 dark:!bg-dark-grey-600 dark:text-dark-grey-900 cursor-not-allowed':
              disabled,
            'bg-white dark:bg-dark-grey-900 text-cool-grey-900 dark:text-cool-grey-100':
              !disabled,
            'border-cool-grey-500/24 dark:border-cool-grey-500/24': !disabled,
          }
        )}
      >
        <span
          className={cn('truncate', {
            'text-cool-grey-500 dark:text-cool-grey-400': !selected,
          })}
        >
          {buttonLabel}
        </span>
        <Icon
          variant="CaretDownIcon"
          className={cn('ml-2 transition-transform flex-shrink-0', {
            'rotate-180': isOpen,
          })}
        />
      </button>

      {isOpen && (
        <TransitionDiv
          isVisible={isOpen}
          className="absolute z-20 left-0 right-0 mt-1 bg-cool-grey-100 dark:bg-dark-grey-800 shadow-sm border rounded py-1 px-2"
        >
          <div className="flex flex-col gap-1">
            <div className="pb-1 mb-1 border-b border-cool-grey-200 dark:border-dark-grey-700">
              <SearchInput
                value={searchQuery}
                onChange={onSearchChange}
                placeholder="Search channels…"
                labelClassName="w-full"
                className="!min-w-0 w-full h-8 text-xs"
                autoFocus
              />
            </div>
            <div
              ref={listRef}
              role="listbox"
              onScroll={handleScroll}
              className="max-h-72 overflow-y-auto overflow-x-hidden flex flex-col gap-1"
            >
              {isLoadingFirstPage ? (
                <div className="flex items-center gap-2 px-2 py-2 text-sm text-cool-grey-500 dark:text-cool-grey-400">
                  <Icon variant="Loading" /> Loading channels…
                </div>
              ) : filtered.length === 0 ? (
                <div className="px-2 py-2 text-sm text-cool-grey-500 dark:text-cool-grey-400">
                  {searchQuery
                    ? hasMore
                      ? 'Searching more channels…'
                      : 'No channels match your search.'
                    : 'No channels available.'}
                </div>
              ) : (
                filtered.map((channel) => {
                  const channelId = channel.id ?? ''
                  const isSelected = channelId === value
                  return (
                    <button
                      key={channelId}
                      type="button"
                      role="option"
                      aria-selected={isSelected}
                      onClick={() => {
                        onChange(channelId, channel.name ?? '')
                        setIsOpen(false)
                      }}
                      className={cn(
                        'transition duration-200 px-2 py-1 -mx-1.5 cursor-pointer select-none rounded text-sm font-sans text-left flex items-center gap-2',
                        {
                          'text-white bg-primary-600': isSelected,
                          'hover:bg-black/5 dark:hover:bg-white/5': !isSelected,
                        }
                      )}
                    >
                      <span className="truncate flex-1">
                        {channel.name ? `#${channel.name}` : channelId}
                      </span>
                      {channel.is_private ? (
                        <Icon variant="LockIcon" size={12} />
                      ) : null}
                    </button>
                  )
                })
              )}
              {isFetchingNextPage ? (
                <div className="flex items-center gap-2 px-2 py-2 text-sm text-cool-grey-500 dark:text-cool-grey-400">
                  <Icon variant="Loading" /> Loading more…
                </div>
              ) : null}
            </div>
            {hasMore && !isFetchingNextPage && filtered.length > 0 ? (
              <Text variant="subtext" theme="neutral" className="px-2 py-1">
                Scroll for more channels
              </Text>
            ) : null}
          </div>
        </TransitionDiv>
      )}
    </div>
  )
}
