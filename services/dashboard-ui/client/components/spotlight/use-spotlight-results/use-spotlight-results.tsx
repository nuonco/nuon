import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { TIconVariant } from '@/components/common/Icon'
import { getOrgs } from '@/lib/ctl-api/orgs/get-orgs'
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
  InstallViewCurrentInputsModal,
  InstallViewStateModal,
  InstallEditStackOverridesModal,
} from '../InstallCommandModals'
import { AppBuildAllComponentsModal } from '../AppCommandModals'
import {
  SpotlightBuildComponentModal,
  SpotlightDeployComponentModal,
  SpotlightTeardownComponentModal,
  SpotlightDriftScanComponentModal,
} from '../ComponentCommandModals'
import { SpotlightRunActionModal } from '../ActionCommandModals'
import { RestartRunnerModalContainer as RestartRunnerModal } from '../RestartRunnerModal'
import {
  type SpotlightResult,
  type ParsedQuery,
  STATIC_PAGES,
  INSTALL_SUB_PAGES,
  APP_SUB_PAGES,
  APP_BRANCH_SUB_PAGES,
  tokenMatch,
} from '../types'

export function useSpotlightResults(
  parsed: ParsedQuery,
  liveParsed: ParsedQuery,
  orgId: string,
  onClose: () => void,
  hasAppBranches = false,
  addModal?: (modal: React.ReactElement) => string,
  orgFeatures?: Record<string, boolean>
) {
  const appSubPages = useMemo(
    () => hasAppBranches ? [...APP_SUB_PAGES, ...APP_BRANCH_SUB_PAGES] : APP_SUB_PAGES,
    [hasAppBranches]
  )

  const { data: appsResult, isFetching: appsFetching } = useQuery({
    queryKey: ['spotlight', 'apps', parsed.query, orgId],
    queryFn: () => getApps({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: (parsed.prefix === 'app' || (parsed.prefix === null && parsed.query.length > 0)) && !!orgId,
  })

  const isRunnerIdQuery = parsed.query.startsWith('run')
  const isGlobalCommand = parsed.prefix === null && parsed.query.startsWith('/')

  const { data: installsResult, isFetching: installsFetching } = useQuery({
    queryKey: ['spotlight', 'installs', parsed.query, orgId],
    queryFn: () =>
      getInstalls({ orgId, q: parsed.query || undefined, limit: 5 }),
    enabled: (parsed.prefix === 'install' || (parsed.prefix === null && parsed.query.length > 0 && !isGlobalCommand)) && !!orgId,
  })

  const { data: globalCommandInstalls, isFetching: globalCommandInstallsFetching } = useQuery({
    queryKey: ['spotlight', 'global-command-installs', orgId],
    queryFn: () => getInstalls({ orgId, limit: 20 }),
    enabled: isGlobalCommand && !!orgId,
  })

  const { data: runnerInstallsResult, isFetching: runnerInstallsFetching } = useQuery({
    queryKey: ['spotlight', 'installs-by-runner', parsed.query, orgId],
    queryFn: () =>
      getInstalls({ orgId, runner_id: parsed.query, limit: 5 }),
    enabled: isRunnerIdQuery && (parsed.prefix === null || parsed.prefix === 'install') && !!orgId,
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
    enabled: parsed.prefix === 'action' && !!orgId,
  })

  const { data: orgsResult, isFetching: orgsFetching } = useQuery({
    queryKey: ['spotlight', 'orgs', parsed.query],
    queryFn: () => getOrgs({ q: parsed.query || undefined, limit: 10 }),
    enabled: parsed.prefix === 'org',
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
    enabled: parsed.prefix === 'component' && !!orgId,
  })

  const results = useMemo((): SpotlightResult[] => {
    if (liveParsed.prefix === null) {
      const pages = STATIC_PAGES.filter((p) => !p.feature || !!orgFeatures?.[p.feature])
      if (!liveParsed.query) return pages
      const matched = pages.filter((p) => tokenMatch(p.label, liveParsed.query))
      const apps = (appsResult?.data ?? []).map((app): SpotlightResult => ({
        label: app.name ?? app.id!,
        tag: 'app',
        path: `/apps/${app.id}`,
        icon: 'AppWindowIcon',
      }))
      const nameInstalls = installsResult?.data ?? []
      const runnerInstalls = runnerInstallsResult?.data ?? []
      const seen = new Set<string>()
      const allInstalls = [...runnerInstalls, ...nameInstalls].filter((i) => {
        if (seen.has(i.id!)) return false
        seen.add(i.id!)
        return true
      })
      const runnerPages: SpotlightResult[] = runnerInstalls.map((install): SpotlightResult => ({
        label: `${install.name ?? install.id!} › Runner`,
        subtitle: install.app?.name,
        tag: 'install',
        path: `/installs/${install.id}/runner`,
        icon: 'CubeIcon',
      }))
      const installs = allInstalls.map((install): SpotlightResult => ({
        label: install.name ?? install.id!,
        subtitle: install.app?.name,
        tag: 'install',
        path: `/installs/${install.id}`,
        icon: 'CubeIcon',
      }))
      if (isGlobalCommand) {
        const commandQuery = liveParsed.query.slice(1).trim().toLowerCase()
        const gcInstalls = globalCommandInstalls?.data ?? []
        const commandResults: SpotlightResult[] = []
        for (const install of gcInstalls) {
          const installId = install.id!
          const name = install.name ?? installId
          const commands: SpotlightResult[] = [
            {
              label: `${name} › Run adhoc action`,
              subtitle: install.app?.name,
              tag: 'command',
              icon: 'TerminalWindowIcon',
              action: () => addModal?.(<InstallAdhocActionModal installId={installId} />),
            },
            {
              label: `${name} › Deploy all components`,
              subtitle: install.app?.name,
              tag: 'command',
              icon: 'LightningIcon',
              action: () => addModal?.(<InstallDeployAllComponentsModal installId={installId} />),
            },
            {
              label: `${name} › Reprovision install`,
              subtitle: install.app?.name,
              tag: 'command',
              icon: 'LightningIcon',
              action: () => addModal?.(<InstallReprovisionModal installId={installId} />),
            },
            {
              label: `${name} › Sync secrets`,
              subtitle: install.app?.name,
              tag: 'command',
              icon: 'LightningIcon',
              action: () => addModal?.(<InstallSyncSecretsModal installId={installId} />),
            },
          ]
          for (const cmd of commands) {
            const cmdName = cmd.label.split(' › ')[1].toLowerCase()
            if (!commandQuery || tokenMatch(cmdName, commandQuery)) {
              commandResults.push(cmd)
            }
          }
        }
        return commandResults
      }

      return [...matched, ...apps, ...runnerPages, ...installs]
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
            icon: 'AppWindowIcon',
          })
          for (const sub of appSubPages) {
            const entry = {
              label: `${name} › ${sub}`,
              tag: 'app',
              path: `/apps/${appId}/${sub.toLowerCase()}`,
              icon: 'AppWindowIcon' as TIconVariant,
            }
            if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
            items.push(entry)
          }
        }
        const commands: SpotlightResult[] = [
          {
            label: `${name} › Build all components`,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<AppBuildAllComponentsModal appId={appId} />),
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
      const nameInstalls = installsResult?.data ?? []
      const runnerInstalls = runnerInstallsResult?.data ?? []
      const seen = new Set<string>()
      const installs = [...runnerInstalls, ...nameInstalls].filter((i) => {
        if (seen.has(i.id!)) return false
        seen.add(i.id!)
        return true
      })
      const runnerIdSet = new Set(runnerInstalls.map((i) => i.id!))
      const items: SpotlightResult[] = []
      for (const install of installs) {
        const installId = install.id!
        const name = install.name ?? installId
        if (parsed.command === null) {
          if (runnerIdSet.has(installId)) {
            items.push({
              label: `${name} › Runner`,
              subtitle: install.app?.name,
              tag: 'install',
              path: `/installs/${installId}/runner`,
              icon: 'CubeIcon',
            })
          }
          items.push({
            label: name,
            subtitle: install.app?.name,
            tag: 'install',
            path: `/installs/${installId}`,
            icon: 'CubeIcon',
          })
          for (const sub of INSTALL_SUB_PAGES) {
            const entry = {
              label: `${name} › ${sub}`,
              subtitle: install.app?.name,
              tag: 'install',
              path: `/installs/${installId}/${sub.toLowerCase()}`,
              icon: 'CubeIcon' as TIconVariant,
            }
            if (parsed.query && !tokenMatch(entry.label, parsed.query)) continue
            items.push(entry)
          }
        }
        const commands: SpotlightResult[] = [
          {
            label: `${name} › Deploy all components`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallDeployAllComponentsModal installId={installId} />),
          },
          {
            label: `${name} › Edit inputs`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallEditInputsModal installId={installId} />),
          },
          {
            label: `${name} › Edit stack overrides`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallEditStackOverridesModal installId={installId} />),
          },
          {
            label: `${name} › Reprovision install`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallReprovisionModal installId={installId} />),
          },
          {
            label: `${name} › Reprovision sandbox`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallReprovisionSandboxModal installId={installId} />),
          },
          ...(install.runner_id ? [{
            label: `${name} › Restart runner`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<RestartRunnerModal runnerId={install.runner_id!} />),
          } satisfies SpotlightResult] : []),
          {
            label: `${name} › Run adhoc action`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallAdhocActionModal installId={installId} />),
          },
          {
            label: `${name} › Sync secrets`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallSyncSecretsModal installId={installId} />),
          },
          {
            label: `${name} › View current inputs`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallViewCurrentInputsModal installId={installId} />),
          },
          {
            label: `${name} › View state`,
            subtitle: install.app?.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<InstallViewStateModal installId={installId} />),
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

    if (parsed.prefix === 'org') {
      return (orgsResult ?? []).map((org): SpotlightResult => ({
        label: org.name ?? org.id!,
        tag: 'org',
        path: `/${org.id}`,
        icon: 'BuildingsIcon',
      }))
    }

    if (parsed.prefix === 'action' && actionResults) {
      const items: SpotlightResult[] = []
      const appActionMap = new Map<string, typeof actionResults.appActions[0]['actions'][0]>()
      for (const { actions } of actionResults.appActions) {
        for (const action of actions) {
          if (action.id) appActionMap.set(action.id, action)
        }
      }
      for (const { app, actions } of actionResults.appActions) {
        for (const action of actions) {
          if (parsed.command === null) {
            items.push({
              label: `${app.name} › ${action.name}`,
              tag: 'action',
              path: `/apps/${app.id}/actions/${action.id}`,
              icon: 'AppWindowIcon',
            })
          }
        }
      }
      for (const { install, actions } of actionResults.installActions) {
        for (const action of actions) {
          const actionWorkflow = action.action_workflow
          const name = actionWorkflow?.name ?? action.action_workflow_id ?? ''
          if (parsed.query && !tokenMatch(name, parsed.query)) continue
          if (parsed.command === null) {
            items.push({
              label: `${install.name} › ${name}`,
              subtitle: install.app?.name,
              tag: 'action',
              path: `/installs/${install.id}/actions/${action.action_workflow_id}`,
              icon: 'CubeIcon',
            })
          }
          const fullAction = appActionMap.get(action.action_workflow_id!) ?? actionWorkflow
          const latestConfig = fullAction?.configs?.at(-1)
          const hasManualTrigger = latestConfig?.triggers?.some((t) => t.type === 'manual')
          if (hasManualTrigger && fullAction && latestConfig) {
            const runCmd: SpotlightResult = {
              label: `${install.name} › ${name} › Run`,
              subtitle: install.app?.name,
              tag: 'command',
              icon: 'LightningIcon',
              action: () => addModal?.(
                <SpotlightRunActionModal
                  installId={install.id!}
                  action={fullAction}
                  actionConfigId={latestConfig.id!}
                />
              ),
            }
            if (parsed.command === null || !parsed.command || tokenMatch('Run', parsed.command)) {
              items.push(runCmd)
            }
          }
        }
      }
      return items
    }

    if (parsed.prefix === 'component' && componentResults) {
      const items: SpotlightResult[] = []
      for (const { app, components } of componentResults.appComps) {
        for (const comp of components) {
          if (parsed.command === null) {
            items.push({
              label: `${app.name} › ${comp.name}`,
              tag: 'component',
              path: `/apps/${app.id}/components/${comp.id}`,
              icon: 'AppWindowIcon',
            })
          }
          const buildCmd: SpotlightResult = {
            label: `${app.name} › ${comp.name} › Build`,
            subtitle: app.name,
            tag: 'command',
            icon: 'LightningIcon',
            action: () => addModal?.(<SpotlightBuildComponentModal appId={app.id!} component={comp} />),
          }
          const cmdName = 'Build'
          if (parsed.command === null || !parsed.command || tokenMatch(cmdName, parsed.command)) {
            items.push(buildCmd)
          }
        }
      }
      for (const { install, components } of componentResults.installComps) {
        for (const comp of components) {
          const compName = comp.component?.name ?? comp.id
          const installId = install.id!
          if (parsed.command === null) {
            items.push({
              label: `${install.name} › ${compName}`,
              tag: 'component',
              path: `/installs/${installId}/components/${comp.component_id}`,
              icon: 'CubeIcon',
            })
          }
          if (comp.component) {
            const component = comp.component
            const commands: SpotlightResult[] = [
              {
                label: `${install.name} › ${compName} › Deploy`,
                subtitle: install.app?.name,
                tag: 'command',
                icon: 'LightningIcon',
                action: () => addModal?.(<SpotlightDeployComponentModal installId={installId} component={component} />),
              },
              {
                label: `${install.name} › ${compName} › Teardown`,
                subtitle: install.app?.name,
                tag: 'command',
                icon: 'LightningIcon',
                action: () => addModal?.(<SpotlightTeardownComponentModal installId={installId} component={component} />),
              },
              {
                label: `${install.name} › ${compName} › Drift scan`,
                subtitle: install.app?.name,
                tag: 'command',
                icon: 'LightningIcon',
                action: () => addModal?.(<SpotlightDriftScanComponentModal installId={installId} component={component} />),
              },
            ]
            for (const cmd of commands) {
              const cmdName = cmd.label.split(' › ').pop()!
              if (parsed.command === null || !parsed.command || tokenMatch(cmdName, parsed.command)) {
                items.push(cmd)
              }
            }
          }
        }
      }
      return items
    }

    return []
  }, [liveParsed, parsed, appsResult, installsResult, runnerInstallsResult, globalCommandInstalls, orgsResult, actionResults, componentResults, appSubPages, addModal, orgFeatures, isGlobalCommand])

  const isFetching = appsFetching || installsFetching || runnerInstallsFetching || globalCommandInstallsFetching || orgsFetching || actionsFetching || componentsFetching

  return { results, isFetching }
}
