import {
  useState,
  useEffect,
  useRef,
  useCallback,
  useMemo,
  type KeyboardEvent,
} from 'react'
import { useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { SearchInput } from '@/components/common/SearchInput'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useOrg } from '@/hooks/use-org'
import { Badge } from '@/components/common/Badge'
import { Skeleton } from '@/components/common/Skeleton'
import { cn } from '@/utils/classnames'
import { getApps } from '@/lib/ctl-api/apps/get-apps'
import { getInstalls } from '@/lib/ctl-api/installs/get-installs'
import { getComponents } from '@/lib/ctl-api/apps/components/get-components'
import { getActions } from '@/lib/ctl-api/apps/actions/get-actions'
import { getInstallActionsLatestRuns } from '@/lib/ctl-api/installs/actions/get-install-actions-latest-runs'
import { getInstallComponents } from '@/lib/ctl-api/installs/components/get-install-components'

type SpotlightResult = {
  label: string
  subtitle?: string
  tag?: string
  path: string
  icon: TIconVariant
}

type ParsedQuery = {
  prefix: 'app' | 'install' | 'component' | 'action' | null
  query: string
}

const STATIC_PAGES: (SpotlightResult & { feature?: string })[] = [
  { label: 'Dashboard', path: '/', icon: 'House', feature: 'org-dashboard' },
  { label: 'Apps', path: '/apps', icon: 'AppWindow' },
  { label: 'Installs', path: '/installs', icon: 'Cube' },
  { label: 'Team', path: '/team', icon: 'UsersThree' },
  { label: 'Build runner', path: '/runner', icon: 'Hammer' },
]

const INSTALL_SUB_PAGES = [
  'Components',
  'Actions',
  'Runner',
  'Workflows',
  'Stacks',
]

const APP_SUB_PAGES = [
  'Components',
  'Actions',
  'Roles',
  'Policies',
  'Installs',
]

const APP_BRANCH_SUB_PAGES = [
  'Branches',
  'Sandbox',
]

const PREFIX_MAP: Record<string, ParsedQuery['prefix']> = {
  'app:': 'app',
  'apps:': 'app',
  'install:': 'install',
  'installs:': 'install',
  'component:': 'component',
  'components:': 'component',
  'action:': 'action',
  'actions:': 'action',
}

function parseQuery(raw: string): ParsedQuery {
  for (const [p, prefix] of Object.entries(PREFIX_MAP)) {
    if (raw.startsWith(p)) {
      return { prefix, query: raw.slice(p.length).trim() }
    }
  }
  return { prefix: null, query: raw.trim() }
}

const FILTER_PREFIXES = ['app:', 'install:', 'component:', 'action:']

function getAutocompletion(input: string): string | null {
  if (!input || input.includes(':')) return null
  const lower = input.toLowerCase()
  const match = FILTER_PREFIXES.find((p) => p.startsWith(lower) && p !== lower)
  return match ?? null
}

function tokenMatch(text: string, query: string): boolean {
  const tokens = query.toLowerCase().split(/\s+/).filter(Boolean)
  const lower = text.toLowerCase()
  return tokens.every((t) => lower.includes(t))
}

interface ISpotlightModal extends IModal {}

export const SpotlightModal = ({ ...props }: ISpotlightModal) => {
  const { removeModal } = useSurfaces()
  const { org } = useOrg()
  const navigate = useNavigate()
  const hasAppBranches = !!org?.features?.['app-branches']
  const appSubPages = useMemo(
    () => hasAppBranches ? [...APP_SUB_PAGES, ...APP_BRANCH_SUB_PAGES] : APP_SUB_PAGES,
    [hasAppBranches]
  )
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

  const orgId = org?.id ?? ''

  const { data: appsResult, isFetching: appsFetching } = useQuery({
    queryKey: ['spotlight', 'apps', parsed.query, orgId],
    queryFn: () => getApps({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: (parsed.prefix === 'app' || (parsed.prefix === null && parsed.query.length > 0)) && !!orgId,
  })

  const { data: installsResult, isFetching: installsFetching } = useQuery({
    queryKey: ['spotlight', 'installs', parsed.query, orgId],
    queryFn: () =>
      getInstalls({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: (parsed.prefix === 'install' || (parsed.prefix === null && parsed.query.length > 0)) && !!orgId,
  })

  const { data: actionResults, isFetching: actionsFetching } = useQuery({
    queryKey: ['spotlight', 'actions', parsed.query, orgId],
    queryFn: async () => {
      const [appsRes, installsRes] = await Promise.all([
        getApps({ orgId, limit: 20 }),
        getInstalls({ orgId, limit: 20 }),
      ])

      const apps = (appsRes.data ?? []).slice(0, 5)
      const installs = (installsRes.data ?? []).slice(0, 5)

      const [appActionResults, installActionResults] = await Promise.all([
        Promise.allSettled(
          apps.map((app) =>
            getActions({
              appId: app.id!,
              orgId,
              q: parsed.query || undefined,
              limit: 3,
            }).then((res) => ({ app, actions: res.data ?? [] }))
          )
        ),
        Promise.allSettled(
          installs.map((install) =>
            getInstallActionsLatestRuns({
              installId: install.id!,
              orgId,
              limit: 20,
            }).then((res) => ({ install, actions: res.data ?? [] }))
          )
        ),
      ])

      const appActions = appActionResults.flatMap((r) => r.status === 'fulfilled' ? [r.value] : [])
      const installActions = installActionResults.flatMap((r) => r.status === 'fulfilled' ? [r.value] : [])

      return { appActions, installActions }
    },
    enabled:
      parsed.prefix === 'action' &&
      !!orgId,
  })

  const { data: componentResults, isFetching: componentsFetching } = useQuery({
    queryKey: ['spotlight', 'components', parsed.query, orgId],
    queryFn: async () => {
      const [appsRes, installsRes] = await Promise.all([
        getApps({ orgId, limit: 20 }),
        getInstalls({ orgId, limit: 20 }),
      ])

      const apps = (appsRes.data ?? []).slice(0, 5)
      const installs = (installsRes.data ?? []).slice(0, 5)

      const [appCompResults, installCompResults] = await Promise.all([
        Promise.allSettled(
          apps.map((app) =>
            getComponents({
              appId: app.id!,
              orgId,
              q: parsed.query || undefined,
              limit: 3,
            }).then((res) => ({ app, components: res.data ?? [] }))
          )
        ),
        Promise.allSettled(
          installs.map((install) =>
            getInstallComponents({
              installId: install.id!,
              orgId,
              q: parsed.query || undefined,
              limit: 3,
            }).then((res) => ({ install, components: res.data ?? [] }))
          )
        ),
      ])

      const appComps = appCompResults.flatMap((r) => r.status === 'fulfilled' ? [r.value] : [])
      const installComps = installCompResults.flatMap((r) => r.status === 'fulfilled' ? [r.value] : [])

      return { appComps, installComps }
    },
    enabled:
      parsed.prefix === 'component' &&
      !!orgId,
  })

  const results = useMemo((): SpotlightResult[] => {
    if (liveParsed.prefix === null) {
      const pages = STATIC_PAGES.filter((p) => !p.feature || !!org?.features?.[p.feature])
      if (!liveParsed.query) return pages
      const matched = pages.filter((p) => tokenMatch(p.label, liveParsed.query))
      const apps = (appsResult?.data ?? []).map((app): SpotlightResult => ({
        label: app.name ?? app.id!,
        tag: 'app',
        path: `/apps/${app.id}`,
        icon: 'AppWindow',
      }))
      const installs = (installsResult?.data ?? []).map((install): SpotlightResult => ({
        label: install.name ?? install.id!,
        subtitle: install.app?.name,
        tag: 'install',
        path: `/installs/${install.id}`,
        icon: 'Cube',
      }))
      return [...matched, ...apps, ...installs]
    }

    if (parsed.prefix === 'app') {
      const apps = appsResult?.data ?? []
      const items: SpotlightResult[] = []
      for (const app of apps) {
        items.push({
          label: app.name ?? app.id!,
          tag: 'app',
          path: `/apps/${app.id}`,
          icon: 'AppWindow',
        })
        for (const sub of appSubPages) {
          const entry = {
            label: `${app.name ?? app.id} › ${sub}`,
            tag: 'app',
            path: `/apps/${app.id}/${sub.toLowerCase()}`,
            icon: 'AppWindow' as TIconVariant,
          }
          if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
          items.push(entry)
        }
      }
      return items
    }

    if (parsed.prefix === 'install') {
      const installs = installsResult?.data ?? []
      const items: SpotlightResult[] = []
      for (const install of installs) {
        items.push({
          label: install.name ?? install.id!,
          subtitle: install.app?.name,
          tag: 'install',
          path: `/installs/${install.id}`,
          icon: 'Cube',
        })
        for (const sub of INSTALL_SUB_PAGES) {
          const entry = {
            label: `${install.name ?? install.id} › ${sub}`,
            subtitle: install.app?.name,
            tag: 'install',
            path: `/installs/${install.id}/${sub.toLowerCase()}`,
            icon: 'Cube' as TIconVariant,
          }
          if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
          items.push(entry)
        }
      }
      return items
    }

    if (parsed.prefix === 'action' && actionResults) {
      const items: SpotlightResult[] = []
      for (const { app, actions } of actionResults.appActions) {
        for (const action of actions) {
          items.push({
            label: `${app.name} › ${action.name}`,
            tag: 'action',
            path: `/apps/${app.id}/actions/${action.id}`,
            icon: 'AppWindow',
          })
        }
      }
      for (const { install, actions } of actionResults.installActions) {
        for (const action of actions) {
          const name = action.action_workflow?.name ?? action.action_workflow_id ?? ''
          if (parsed.query && !tokenMatch(name, parsed.query)) continue
          items.push({
            label: `${install.name} › ${name}`,
            subtitle: install.app?.name,
            tag: 'action',
            path: `/installs/${install.id}/actions/${action.action_workflow_id}`,
            icon: 'Cube',
          })
        }
      }
      return items
    }

    if (parsed.prefix === 'component' && componentResults) {
      const items: SpotlightResult[] = []
      for (const { app, components } of componentResults.appComps) {
        for (const comp of components) {
          items.push({
            label: `${app.name} › ${comp.name}`,
            tag: 'component',
            path: `/apps/${app.id}/components/${comp.id}`,
            icon: 'AppWindow',
          })
        }
      }
      for (const { install, components } of componentResults.installComps) {
        for (const comp of components) {
          items.push({
            label: `${install.name} › ${comp.component?.name ?? comp.id}`,
            tag: 'component',
            path: `/installs/${install.id}/components/${comp.component_id}`,
            icon: 'Cube',
          })
        }
      }
      return items
    }

    return []
  }, [liveParsed, parsed, appsResult, installsResult, actionResults, componentResults, appSubPages])

  const isSearching = raw !== debouncedRaw || appsFetching || installsFetching || actionsFetching || componentsFetching

  useEffect(() => {
    setActiveIndex(0)
  }, [raw])

  const close = useCallback(() => {
    removeModal(props.modalId)
  }, [removeModal, props.modalId])

  const selectResult = useCallback(
    (result: SpotlightResult) => {
      navigate(`/${orgId}${result.path}`)
      close()
    },
    [navigate, orgId, close]
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
    const active = listRef.current?.querySelector(`[data-index="${activeIndex}"]`) as HTMLElement
    active?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  return (
    <Modal
      size="half"
      showHeader={false}
      showFooter={false}
      {...props}
      className="!mt-[15vh] !mb-auto"
      childrenClassName="!p-0 !gap-0"
    >
      <div ref={inputWrapperRef} className="p-4 border-b" onKeyDown={handleKeyDown}>
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
            <div className="absolute inset-0 pointer-events-none flex items-center pl-8 pr-3.5 text-sm text-cool-grey-500 dark:text-cool-grey-500">
              <span className="invisible">{raw}</span>
              <span>{autocompletion.slice(raw.length)}</span>
              <span className="ml-1.5 text-xs text-cool-grey-500 dark:text-cool-grey-500 border border-cool-grey-400 dark:border-dark-grey-500 rounded px-1">tab</span>
            </div>
          )}
        </div>
      </div>
      <div className="px-2 py-1">
        {liveParsed.prefix === null && (
          <div className="px-2 py-1 flex items-center gap-1.5 flex-wrap">
            <Text variant="subtext" className="text-cool-grey-600">Filter by</Text>
            {FILTER_PREFIXES.map((prefix) => (
              <button key={prefix} onClick={() => setRaw(prefix)} className="cursor-pointer">
                <Badge size="sm" variant="code" theme="neutral">{prefix}</Badge>
              </button>
            ))}
          </div>
        )}
      </div>
      <div ref={listRef} className="max-h-72 overflow-y-auto py-1 px-2">
        <div className="flex flex-col gap-1">
          {results.length === 0 && raw.length > 0 && isSearching && (
            <div className="flex flex-col gap-2 px-2 py-2">
              <Skeleton width={['70%', '55%', '40%']} lines={3} height="1.5rem" />
            </div>
          )}
          {results.length === 0 && raw.length > 0 && !isSearching && (
            <div className="px-2 py-2 text-sm text-cool-grey-700 dark:text-cool-grey-400 flex flex-col gap-1">
              <span>No results for &ldquo;{raw}&rdquo;</span>
              <span className="text-xs text-cool-grey-600 dark:text-cool-grey-500">
                Try{' '}
                <button className="underline cursor-pointer" onClick={() => setRaw(`app:${liveParsed.query} `)}>app:</button>
                {' '}<button className="underline cursor-pointer" onClick={() => setRaw(`install:${liveParsed.query} `)}>install:</button>
                {' '}<button className="underline cursor-pointer" onClick={() => setRaw(`component:${liveParsed.query} `)}>component:</button>
                {' '}<button className="underline cursor-pointer" onClick={() => setRaw(`action:${liveParsed.query} `)}>action:</button>
                {' '}to narrow your search
              </span>
            </div>
          )}
          {results.map((result, i) => (
            <button
              key={result.path}
              data-index={i}
              className={cn(
                'transition duration-200 px-2 py-1 -mx-1.5 cursor-pointer select-none rounded text-sm text-left flex items-center gap-3',
                {
                  'text-white bg-primary-600': i === activeIndex,
                  'hover:bg-black/5 dark:hover:bg-white/5': i !== activeIndex,
                }
              )}
              onClick={() => selectResult(result)}
              onMouseEnter={() => setActiveIndex(i)}
            >
              <Icon
                variant={result.icon}
                className={cn('shrink-0', {
                  'text-white': i === activeIndex,
                  'text-cool-grey-700 dark:text-cool-grey-500': i !== activeIndex,
                })}
              />
              <div className="flex flex-col min-w-0 flex-1">
                <span className="truncate">{result.label}</span>
                {result.subtitle && (
                  <span
                    className={cn('text-xs truncate', {
                      'text-white/70': i === activeIndex,
                      'text-cool-grey-500': i !== activeIndex,
                    })}
                  >
                    {result.subtitle}
                  </span>
                )}
              </div>
              {result.tag && (
                <Badge size="sm" variant="code" theme="neutral" className="shrink-0">
                  {result.tag}
                </Badge>
              )}
            </button>
          ))}
        </div>
      </div>
    </Modal>
  )
}
