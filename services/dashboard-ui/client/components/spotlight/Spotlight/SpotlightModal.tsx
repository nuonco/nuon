import {
  useState,
  useEffect,
  useRef,
  useCallback,
  useMemo,
  type KeyboardEvent,
} from 'react'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Badge } from '@/components/common/Badge'
import { Skeleton } from '@/components/common/Skeleton'
import { FILTER_PREFIXES, COMMANDS_BY_PREFIX, parseQuery, getAutocompletion } from '../types'
import { useSpotlightResults } from '../use-spotlight-results'
import { SpotlightResultItem } from '../SpotlightResultItem'
import type { SpotlightResult } from '../types'

interface ISpotlightModal extends IModal {
  orgId: string
  onClose: () => void
  onNavigate: (path: string) => void
  onAddModal?: (modal: React.ReactElement) => string
  orgFeatures?: Record<string, boolean>
}

export const SpotlightModal = ({ orgId, onClose, onNavigate, onAddModal, orgFeatures, ...props }: ISpotlightModal) => {
  const [raw, setRaw] = useState('')
  const [debouncedRaw, setDebouncedRaw] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const autocompletion = useMemo(() => getAutocompletion(raw), [raw])
  const listRef = useRef<HTMLDivElement>(null)
  const inputWrapperRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const t = setTimeout(() => {
      inputWrapperRef.current?.querySelector('input')?.focus()
    }, 200)
    return () => clearTimeout(t)
  }, [])

  useEffect(() => {
    const t = setTimeout(() => setDebouncedRaw(raw), 300)
    return () => clearTimeout(t)
  }, [raw])

  const parsed = useMemo(() => parseQuery(debouncedRaw), [debouncedRaw])
  const liveParsed = useMemo(() => parseQuery(raw), [raw])

  const { results, isFetching } = useSpotlightResults(parsed, liveParsed, orgId, onClose, false, onAddModal, orgFeatures)
  const isSearching = raw !== debouncedRaw || isFetching

  useEffect(() => {
    setActiveIndex(0)
  }, [raw])

  const selectResult = useCallback(
    (result: SpotlightResult) => {
      if (result.action) {
        const action = result.action
        onClose()
        requestAnimationFrame(() => action())
      } else if (result.tag === 'org') {
        onClose()
        onNavigate(result.path!)
      } else {
        onClose()
        onNavigate(`/${orgId}${result.path}`)
      }
    },
    [onNavigate, orgId, onClose]
  )

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLDivElement>) => {
      if (e.key === 'Tab' && autocompletion) {
        e.preventDefault()
        setRaw(autocompletion)
      } else if (e.key === 'ArrowDown') {
        e.preventDefault()
        setActiveIndex((i) => Math.min(i + 1, results.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setActiveIndex((i) => Math.max(i - 1, 0))
      } else if (e.key === 'Enter' && results[activeIndex]) {
        e.preventDefault()
        selectResult(results[activeIndex])
      }
    },
    [results, activeIndex, selectResult, autocompletion]
  )

  useEffect(() => {
    const active = listRef.current?.querySelector(
      `[data-index="${activeIndex}"]`
    ) as HTMLElement
    active?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  return (
    <Modal
      size="lg"
      showHeader={false}
      showFooter={false}
      {...props}
      className="!mt-[15vh] !mb-auto"
      childrenClassName="!p-0 !gap-0"
    >
      <div
        ref={inputWrapperRef}
        className="p-4 border-b"
        onKeyDown={handleKeyDown}
      >
        <div className="relative">
          <SearchInput
            className="w-full bg-transparent"
            labelClassName="w-full"
            placeholder="Search pages, apps, installs, components, actions…"
            value={raw}
            onChange={setRaw}
            onClear={() => setRaw('')}
            autoFocus
          />
          {autocompletion && (
            <div className="absolute inset-0 pointer-events-none flex items-center pl-8 pr-3.5 text-sm text-cool-grey-500 dark:text-cool-grey-500 whitespace-pre">
              <span className="invisible">{raw}</span>
              <span>{autocompletion.slice(raw.length)}</span>
              <span className="ml-1.5 text-xs text-cool-grey-500 dark:text-cool-grey-500 border border-cool-grey-400 dark:border-dark-grey-500 rounded px-1">
                tab
              </span>
            </div>
          )}
        </div>
      </div>
      <div className="px-2 py-1">
        {liveParsed.prefix === null && (
          <div className="px-2 py-1 flex items-center gap-1.5 flex-wrap">
            <Text variant="subtext" className="text-cool-grey-600">
              Filter by
            </Text>
            {FILTER_PREFIXES.map((prefix) => (
              <button
                key={prefix}
                onClick={() => setRaw(prefix)}
                className="cursor-pointer"
              >
                <Badge size="sm" variant="code" theme="neutral">
                  {prefix}
                </Badge>
              </button>
            ))}
          </div>
        )}
        {liveParsed.prefix && COMMANDS_BY_PREFIX[liveParsed.prefix] && liveParsed.command === null && (
          <div className="px-2 py-1 flex items-center gap-1.5">
            <Text variant="subtext" className="text-cool-grey-600">
              Type{' '}
              <Badge size="sm" variant="code" theme="neutral">
                /
              </Badge>{' '}
              to run commands
            </Text>
          </div>
        )}
      </div>
      <div ref={listRef} className="max-h-72 overflow-y-auto py-1 px-2">
        <div className="flex flex-col gap-1">
          {results.length === 0 && raw.length > 0 && isSearching && (
            <>
              <div className="flex items-center gap-3 px-1 py-2">
                <Skeleton width="20px" height="20px" />
                <Skeleton width="110px" height="20px" />
              </div>
              <div className="flex items-center gap-3 px-1 py-2">
                <Skeleton width="20px" height="20px" />
                <Skeleton width="190px" height="20px" />
              </div>
              <div className="flex items-center gap-3 px-1 py-2">
                <Skeleton width="20px" height="20px" />
                <Skeleton width="80px" height="20px" />
              </div>
            </>
          )}
          {results.length === 0 && raw.length > 0 && !isSearching && (
            <div className="px-2 py-2 text-sm text-cool-grey-700 dark:text-cool-grey-400 flex flex-col gap-1">
              <span>No results for &ldquo;{raw}&rdquo;</span>
              <span className="text-xs text-cool-grey-600 dark:text-cool-grey-500">
                Try{' '}
                <button
                  className="underline cursor-pointer"
                  onClick={() => setRaw(`app:${liveParsed.query} `)}
                >
                  app:
                </button>{' '}
                <button
                  className="underline cursor-pointer"
                  onClick={() => setRaw(`install:${liveParsed.query} `)}
                >
                  install:
                </button>{' '}
                <button
                  className="underline cursor-pointer"
                  onClick={() => setRaw(`component:${liveParsed.query} `)}
                >
                  component:
                </button>{' '}
                <button
                  className="underline cursor-pointer"
                  onClick={() => setRaw(`action:${liveParsed.query} `)}
                >
                  action:
                </button>{' '}
                <button
                  className="underline cursor-pointer"
                  onClick={() => setRaw(`org:${liveParsed.query} `)}
                >
                  org:
                </button>{' '}
                to narrow your search
              </span>
            </div>
          )}
          {results.map((result, i) => (
            <SpotlightResultItem
              key={result.path ?? result.label}
              result={result}
              index={i}
              isActive={i === activeIndex}
              onSelect={() => selectResult(result)}
              onHover={() => setActiveIndex(i)}
            />
          ))}
        </div>
      </div>
    </Modal>
  )
}
