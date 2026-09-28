import {
  createApp,
  createAppBranch,
  createAppInstall,
  createBranchConfig,
  getAppBranches,
  getAppConfig,
  getAppConfigs,
  getApp,
  getApps,
  getBranchLatestConfig,
  getBranchWorkflowRuns,
  updateApp,
  type TCreateAppInstallBody,
} from '@/lib'
import type { TApp, TAPIError, TAppConfig, TInstall, TInstallWorkflow } from '@/types'
import {
  CONFIG_DIRECTORY,
  DEFAULT_BRANCH_NAME,
  INSTALL_NAME_ATTEMPTS,
  KITCHEN_SINK_REPO,
  KITCHEN_SINK_VARIANTS,
  TRACKED_GIT_BRANCH,
  installNameFor,
  type TCloud,
  type TExampleCloud,
} from './constants'

const statusOf = (error: unknown) => (error as TAPIError | undefined)?.status

export type TBranchSource =
  | { kind: 'github'; vcsConnectionId: string; repo: string; directory?: string }
  | { kind: 'public'; repo: string; directory?: string }

export async function findAppByName({
  orgId,
  name,
}: {
  orgId: string
  name: string
}): Promise<TApp | undefined> {
  const { data } = await getApps({ orgId, q: name, limit: 50 })
  return (data ?? []).find((app) => app.name === name)
}

// Finds or creates the `default` branch, then gives it a config if it has none.
// No install groups are sent, so ctl-api creates the default one.
export async function ensureDefaultBranch({
  orgId,
  appId,
  source,
}: {
  orgId: string
  appId: string
  source: TBranchSource
}): Promise<string> {
  const { data: branches } = await getAppBranches({
    appId,
    orgId,
    q: DEFAULT_BRANCH_NAME,
    limit: 50,
  })
  const branch =
    (branches ?? []).find((b) => b.name === DEFAULT_BRANCH_NAME) ??
    (await createAppBranch({ appId, orgId, body: { name: DEFAULT_BRANCH_NAME } }))
  const branchId = branch.id as string

  try {
    await getBranchLatestConfig({ appId, branchId, orgId })
    return branchId
  } catch (error) {
    if (statusOf(error) !== 404) throw error
  }

  const vcs = {
    repo: source.repo,
    directory: source.directory ?? CONFIG_DIRECTORY,
    branch: TRACKED_GIT_BRANCH,
  }
  await createBranchConfig({
    appId,
    branchId,
    orgId,
    request:
      source.kind === 'github'
        ? { connected_github_vcs_config: { ...vcs, vcs_connection_id: source.vcsConnectionId } }
        : { public_git_vcs_config: vcs },
  })
  return branchId
}

export class AppNameTakenError extends Error {
  constructor(name: string) {
    super(`An app template named ${name} exists`)
  }
}

export interface ISavedApp {
  appId?: string
  branchId?: string
}

// Each create is skipped when the journey already holds its ID, and progress is
// reported after every call so a failure part way through never repeats one.
export async function setUpOwnApp({
  orgId,
  name,
  repo,
  vcsConnectionId,
  saved,
  onProgress,
}: {
  orgId: string
  name: string
  repo: string
  vcsConnectionId: string
  saved: ISavedApp
  onProgress: (progress: { appId: string; branchId?: string }) => Promise<void>
}): Promise<{ appId: string; branchId: string }> {
  // Back then Next reuses the app; renaming it on the way makes a new one.
  let appId: string | undefined
  if (saved.appId) {
    try {
      const existing = await getApp({ orgId, appId: saved.appId })
      if (existing.name === name) appId = existing.id as string
    } catch (error) {
      if (statusOf(error) !== 404) throw error
    }
  }
  const reuse = !!appId
  if (!appId) {
    try {
      appId = (await createApp({ orgId, body: { name } })).id as string
    } catch (error) {
      if (statusOf(error) === 409) throw new AppNameTakenError(name)
      throw error
    }
    await onProgress({ appId })
  }

  await updateApp({
    appId,
    orgId,
    body: { config_repo: repo, config_directory: CONFIG_DIRECTORY },
  })

  const branchId =
    reuse && saved.branchId
      ? saved.branchId
      : await ensureDefaultBranch({
          orgId,
          appId,
          source: { kind: 'github', vcsConnectionId, repo },
        })
  await onProgress({ appId, branchId })
  return { appId, branchId }
}

export async function setUpKitchenSink({
  orgId,
  cloud,
}: {
  orgId: string
  cloud: TExampleCloud
}): Promise<{ appId: string; appName: string; branchId: string }> {
  const { appName, directory } = KITCHEN_SINK_VARIANTS[cloud]
  let app = await findAppByName({ orgId, name: appName })
  if (!app) {
    try {
      app = await createApp({ orgId, body: { name: appName } })
    } catch (error) {
      // Created by another tab between the lookup and the create.
      if (statusOf(error) !== 409) throw error
      app = await findAppByName({ orgId, name: appName })
      if (!app) throw error
    }
  }
  const appId = app.id as string
  const branchId = await ensureDefaultBranch({
    orgId,
    appId,
    source: { kind: 'public', repo: KITCHEN_SINK_REPO, directory },
  })
  return { appId, appName, branchId }
}

// A push shows up as a git run on the branch. The automatic run ctl-api starts
// when the branch config is created is a manual run, so it never counts.
export const findPushRun = (runs?: TInstallWorkflow[]) =>
  (runs ?? []).find((run) => {
    const meta = (run.metadata ?? {}) as Record<string, unknown>
    const runType = meta.run_type ?? (run as Record<string, unknown>).run_type
    const trigger = meta.event_type ?? meta.trigger
    return runType === 'git-run' && trigger === 'push'
  })

export const pushRunSha = (run?: TInstallWorkflow) => {
  const meta = (run?.metadata ?? {}) as Record<string, unknown>
  const sha = (meta.head_sha ?? meta.commit_sha) as string | undefined
  return sha ? sha.slice(0, 7) : undefined
}

export const getBranchRuns = ({
  orgId,
  appId,
  branchId,
}: {
  orgId: string
  appId: string
  branchId: string
}) => getBranchWorkflowRuns({ orgId, appId, branchId, limit: 20 }).then((res) => res.data ?? [])

// Install create uses the branch's newest active app config, so that is what the
// Deploy step waits for.
export async function findActiveAppConfig({
  orgId,
  appId,
  branchId,
}: {
  orgId: string
  appId: string
  branchId: string
}): Promise<TAppConfig | undefined> {
  const configs = await getAppConfigs({ orgId, appId, limit: 20 })
  const active = (configs ?? []).filter(
    (config) => config.status === 'active' && config.labels?.source !== 'git-preview-run'
  )
  return active.find((config) => config.app_branch_id === branchId) ?? active[0]
}

export const getAppConfigWithInputs = ({
  orgId,
  appId,
  appConfigId,
}: {
  orgId: string
  appId: string
  appConfigId: string
}) => getAppConfig({ orgId, appId, appConfigId, recurse: true })

// First-run installs get no inputs form: every required input the vendor owns
// is sent with its default. Customer inputs are collected by the install stack.
export function defaultInputs(config?: TAppConfig): {
  inputs: Record<string, string>
  missing: string[]
} {
  const inputs: Record<string, string> = {}
  const missing: string[] = []
  for (const input of config?.input?.inputs ?? []) {
    if (!input.required || input.source === 'customer' || !input.name) continue
    if (input.default) inputs[input.name] = input.default
    else missing.push(input.name)
  }
  return { inputs, missing }
}

export const cloudAccountBlock = (
  cloud: TCloud,
  region: string
): Pick<TCreateAppInstallBody, 'aws_account' | 'gcp_account' | 'azure_account'> => {
  switch (cloud) {
    case 'aws':
      return { aws_account: { region, iam_role_arn: '' } }
    case 'gcp':
      return { gcp_account: { region } }
    case 'azure':
      return {
        azure_account: {
          location: region,
          service_principal_app_id: '',
          service_principal_password: '',
          subscription_tenant_id: '',
        },
      }
  }
}

export async function createTestInstall({
  orgId,
  appId,
  appName,
  branchId,
  cloud,
  region,
  autoApprove,
  inputs,
}: {
  orgId: string
  appId: string
  appName: string
  branchId: string
  cloud: TCloud
  region: string
  autoApprove: boolean
  inputs: Record<string, string>
}): Promise<TInstall> {
  let lastError: unknown
  for (let attempt = 0; attempt < INSTALL_NAME_ATTEMPTS; attempt++) {
    try {
      const { data } = await createAppInstall({
        appId,
        orgId,
        body: {
          name: installNameFor(appName, attempt),
          ...cloudAccountBlock(cloud, region),
          app_branch_id: branchId,
          install_config: { approval_option: autoApprove ? 'approve-all' : 'prompt' },
          inputs,
          metadata: { managed_by: 'nuon/dashboard' },
        },
      })
      return data
    } catch (error) {
      if (statusOf(error) !== 409) throw error
      lastError = error
    }
  }
  throw lastError
}
