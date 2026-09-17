import { useRef, type ReactNode } from 'react'
import { Spinner } from '../atoms/Spinner'
import { Text } from '../atoms/Text'
import { ListSearch } from './ListSearch'
import { Menu, MenuItem, MenuSeparator } from './Menu'

export interface ISwitcherMenuItem {
  id: string
  label: string
  href: string
  content?: ReactNode
}

const LOADING_ROWS = 5

const LOADING_ROW_CLASSES =
  'flex min-h-8 items-center gap-2 rounded-md border border-transparent px-2'

export interface ISwitcherMenu {
  items: ISwitcherMenuItem[]
  selectedId?: string
  search: string
  onSearchChange: (value: string) => void
  onLoadMore: () => void
  searchLabel: string
  searchPlaceholder: string
  emptyTitle: string
  errorTitle: string
  loadingContent?: ReactNode
  loading?: boolean
  loadingMore?: boolean
  hasMore?: boolean
  hasError?: boolean
}

export const SwitcherMenu = ({
  items,
  selectedId,
  search,
  onSearchChange,
  onLoadMore,
  searchLabel,
  searchPlaceholder,
  emptyTitle,
  errorTitle,
  loadingContent,
  loading = false,
  loadingMore = false,
  hasMore = false,
  hasError = false,
}: ISwitcherMenu) => {
  const searchRef = useRef<HTMLInputElement>(null)

  return (
    <Menu initialFocusRef={searchRef} className="w-72">
      <div className="pb-1">
        <ListSearch
          ref={searchRef}
          value={search}
          onValueChange={onSearchChange}
          aria-label={searchLabel}
          placeholder={searchPlaceholder}
          inputClassName="rounded-md"
        />
      </div>

      {loading ? (
        <div className="flex flex-col gap-0.5" aria-hidden>
          {Array.from({ length: LOADING_ROWS }, (_, index) => (
            <div key={index} className={LOADING_ROW_CLASSES}>
              {loadingContent ?? <Text loading loadingWidth={14} />}
            </div>
          ))}
        </div>
      ) : hasError ? (
        <Text
          as="div"
          variant="caption"
          color="secondary"
          className="px-2 py-4 text-center"
        >
          {errorTitle}
        </Text>
      ) : items.length ? (
        items.map((item) => (
          <MenuItem
            key={item.id}
            href={item.href}
            selected={item.id === selectedId}
          >
            {item.content ?? (
              <Text family="mono" className="truncate">
                {item.label}
              </Text>
            )}
          </MenuItem>
        ))
      ) : (
        <Text
          as="div"
          variant="caption"
          color="tertiary"
          className="px-2 py-4 text-center"
        >
          {emptyTitle}
        </Text>
      )}

      {hasMore ? (
        <>
          <MenuSeparator />
          <MenuItem
            closeOnSelect={false}
            disabled={loadingMore}
            onSelect={onLoadMore}
            icon={loadingMore ? <Spinner size={14} /> : undefined}
          >
            {loadingMore ? 'Loading more' : 'Load more'}
          </MenuItem>
        </>
      ) : null}
    </Menu>
  )
}
