import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { TIconVariant } from '@/components/common/Icon'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getApps } from '@/lib/ctl-api/apps/get-apps'
import { getInstalls } from '@/lib/ctl-api/installs/get-installs'
import { getComponents } from '@/lib/ctl-api/apps/components/get-components'
import { getActions } from '@/lib/ctl-api/apps/actions/get-actions'
import { getInstallActionsLatestRuns } from '@/lib/ctl-api/installs/actions/get-install-actions-latest-runs'
import { getInstallComponents } from '@/lib/ctl-api/installs/components/get-install-components'
import {
  InstallAdhocActionModal,
  InstallEditInputsModal,
  InstallSyncSecretsModal,
  InstallReprovisionModal,
  InstallReprovisionSandboxModal,
  InstallDeployAllComponentsModal,
} from './InstallCommandModals'
import { AppBuildAllComponentsModal } from './AppCommandModals'
import { RestartRunnerModal } from './RestartRunnerModal'
import {
  type SpotlightResult,
  type ParsedQuery,
  STATIC_PAGES,
  INSTALL_SUB_PAGES,
  APP_SUB_PAGES,
  APP_BRANCH_SUB_PAGES,
  tokenMatch,
} from './types'

export function useSpotlightResults(parsed: ParsedQuery, liveParsed: ParsedQuery) {
  const { org } = useOrg()
  const { addModal } = useSurfaces()
  const orgId = org?.id ?? ''

  const hasAppBranches = !!org?.features?.['app-branches']
  const appSubPages = useMemo(
    () => hasAppBranches ? [...APP_SUB_PAGES, ...APP_BRANCH_SUB_PAGES] : APP_SUB_PAGES,
    [hasAppBranches]
  )

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
        const appId = app.id!
        const name = app.name ?? appId
        if (parsed.command === null) {
          items.push({
            label: name,
            tag: 'app',
            path: `/apps/${appId}`,
            icon: 'AppWindow',
          })
          for (const sub of appSubPages) {
            const entry = {
              label: `${name} › ${sub}`,
              tag: 'app',
              path: `/apps/${appId}/${sub.toLowerCase()}`,
              icon: 'AppWindow' as TIconVariant,
            }
            if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
            items.push(entry)
          }
        }
        const commands: SpotlightResult[] = [
          {
            label: `${name} › Build all components`,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<AppBuildAllComponentsModal appId={appId} />),
          },
        ]
        for (const cmd of commands) {
          const cmdName = cmd.label.split(' › ')[1]
          if (parsed.command === null || !parsed.command || tokenMatch(cmdName, parsed.command)) {
            items.push(cmd)
          }
        }
      }
      return items
    }

    if (parsed.prefix === 'install') {
      const installs = installsResult?.data ?? []
      const items: SpotlightResult[] = []
      for (const install of installs) {
        const installId = install.id!
        const name = install.name ?? installId
        if (parsed.command === null) {
          items.push({
            label: name,
            subtitle: install.app?.name,
            tag: 'install',
            path: `/installs/${installId}`,
            icon: 'Cube',
          })
          for (const sub of INSTALL_SUB_PAGES) {
            const entry = {
              label: `${name} › ${sub}`,
              subtitle: install.app?.name,
              tag: 'install',
              path: `/installs/${installId}/${sub.toLowerCase()}`,
              icon: 'Cube' as TIconVariant,
            }
            if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
            items.push(entry)
          }
        }
        const commands: SpotlightResult[] = [
          {
            label: `${name} › Run adhoc action`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallAdhocActionModal installId={installId} />),
          },
          {
            label: `${name} › Edit inputs`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallEditInputsModal installId={installId} />),
          },
          {
            label: `${name} › Sync secrets`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallSyncSecretsModal installId={installId} />),
          },
          {
            label: `${name} › Reprovision install`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallReprovisionModal installId={installId} />),
          },
          {
            label: `${name} › Reprovision sandbox`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallReprovisionSandboxModal installId={installId} />),
          },
          {
            label: `${name} › Deploy all components`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<InstallDeployAllComponentsModal installId={installId} />),
          },
          ...(install.runner_id ? [{
            label: `${name} › Restart runner`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'Lightning',
            action: () => addModal(<RestartRunnerModal runnerId={install.runner_id!} />),
          } satisfies SpotlightResult] : []),
        ]
        for (const cmd of commands) {
          const cmdName = cmd.label.split(' › ')[1]
          if (parsed.command === null || !parsed.command || tokenMatch(cmdName, parsed.command)) {
            items.push(cmd)
          }
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
  }, [liveParsed, parsed, appsResult, installsResult, actionResults, componentResults, appSubPages, addModal])

  const isFetching = appsFetching || installsFetching || actionsFetching || componentsFetching

  return { results, isFetching }
}
