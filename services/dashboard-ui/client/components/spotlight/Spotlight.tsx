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

  const { data: appsResult } = useQuery({
    queryKey: ['spotlight', 'apps', parsed.query, orgId],
    queryFn: () => getApps({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: parsed.prefix === 'app' && !!orgId,
  })

  const { data: installsResult } = useQuery({
    queryKey: ['spotlight', 'installs', parsed.query, orgId],
    queryFn: () =>
      getInstalls({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: parsed.prefix === 'install' && !!orgId,
  })

  const { data: actionResults } = useQuery({
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
              q: parsed.query || undefined,
              limit: 3,
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
      !!orgId &&
      parsed.query.length > 0,
  })

  const { data: componentResults } = useQuery({
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
      !!orgId &&
      parsed.query.length > 0,
  })

  const results = useMemo((): SpotlightResult[] => {
    if (liveParsed.prefix === null) {
      const pages = STATIC_PAGES.filter((p) => !p.feature || !!org?.features?.[p.feature])
      if (!liveParsed.query) return pages
      return pages.filter((p) => tokenMatch(p.label, liveParsed.query))
    }

    if (parsed.prefix === 'app') {
      const apps = appsResult?.data ?? []
      const items: SpotlightResult[] = []
      for (const app of apps) {
        items.push({
          label: app.name ?? app.id!,
          path: `/apps/${app.id}`,
          icon: 'AppWindow',
        })
        for (const sub of appSubPages) {
          const entry = {
            label: `${app.name ?? app.id} › ${sub}`,
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
          path: `/installs/${install.id}`,
          icon: 'Cube',
        })
        for (const sub of INSTALL_SUB_PAGES) {
          const entry = {
            label: `${install.name ?? install.id} › ${sub}`,
            subtitle: install.app?.name,
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
            path: `/apps/${app.id}/actions/${action.id}`,
            icon: 'TerminalWindow',
          })
        }
      }
      for (const { install, actions } of actionResults.installActions) {
        for (const action of actions) {
          items.push({
            label: `${install.name} › ${action.action_workflow?.name ?? action.action_workflow_id}`,
            subtitle: install.app?.name,
            path: `/installs/${install.id}/actions/${action.action_workflow_id}`,
            icon: 'TerminalWindow',
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
            path: `/apps/${app.id}/components/${comp.id}`,
            icon: 'AppWindow',
          })
        }
      }
      for (const { install, components } of componentResults.installComps) {
        for (const comp of components) {
          items.push({
            label: `${install.name} › ${comp.component?.name ?? comp.id}`,
            path: `/installs/${install.id}/components/${comp.component_id}`,
            icon: 'Cube',
          })
        }
      }
      return items
    }

    return []
  }, [liveParsed, parsed, appsResult, installsResult, actionResults, componentResults, appSubPages])

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
      if (e.key === 'ArrowDown') {
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
    [results, activeIndex, selectResult]
  )

  useEffect(() => {
    const active = listRef.current?.children[activeIndex] as HTMLElement
    active?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  return (
    <Modal
      size="half"
      showHeader={false}
      showFooter={false}
      {...props}
      childrenClassName="!p-0 !gap-0"
    >
      <div ref={inputWrapperRef} className="p-4 border-b" onKeyDown={handleKeyDown}>
        <SearchInput
        className="w-full"
          labelClassName="w-full"
          placeholder="Search pages, apps, installs, components, actions…"
          value={raw}
          onChange={setRaw}
          onClear={() => setRaw('')}
          autoFocus
        />
      </div>
      <div className="px-2 py-1">
        {liveParsed.prefix === null && (
          <div className="px-2 py-1">
            <Text variant="subtext" className="text-cool-grey-600">
              Type{' '}
              <code className="text-xs bg-cool-grey-100 dark:bg-dark-grey-800 px-1 rounded">
                app:
              </code>{' '}
              <code className="text-xs bg-cool-grey-100 dark:bg-dark-grey-800 px-1 rounded">
                install:
              </code>{' '}
              <code className="text-xs bg-cool-grey-100 dark:bg-dark-grey-800 px-1 rounded">
                component:
              </code>{' '}
              or{' '}
              <code className="text-xs bg-cool-grey-100 dark:bg-dark-grey-800 px-1 rounded">
                action:
              </code>{' '}
              to search entities
            </Text>
          </div>
        )}
      </div>
      <div ref={listRef} className="max-h-72 overflow-y-auto py-1 px-2">
        <div className="flex flex-col gap-1">
          {results.length === 0 && raw.length > 0 && (
            <div className="px-2 py-1 text-sm text-cool-grey-500 dark:text-cool-grey-400">
              No results
            </div>
          )}
          {results.map((result, i) => (
            <button
              key={result.path}
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
                  'text-cool-grey-500': i !== activeIndex,
                })}
              />
              <div className="flex flex-col min-w-0">
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
            </button>
          ))}
        </div>
      </div>
    </Modal>
  )
}
