import {
  FIRST_RUN_STEPS,
  firstIncompleteStep,
  type TFirstRunMetadata,
  type TFirstRunStep,
} from '@/hooks/use-first-run-journey'
import { createTrialOrg, getApp, getInstall, getOrgs } from '@/lib'
import { getOrgSession, setOrgSession } from '@/lib/cookies'
import { trackEvent } from '@/lib/posthog-analytics'
import type { IUser, TAPIError, TOrg, TUserJourney } from '@/types'
import { defaultRegion, isCloud, type TCloud, type TPath } from './constants'

const statusOf = (error: unknown) => (error as TAPIError | undefined)?.status

// A new sign-up has no org yet: onboarding creates one before anything renders.
// Anyone else keeps working in their current org.
export async function resolveFirstRunOrg({ user }: { user: IUser | null }): Promise<TOrg> {
  const orgs = await getOrgs({ limit: 50 })
  let org: TOrg | undefined
  if (orgs?.length) {
    const session = getOrgSession()
    org = orgs.find((o) => o.id === session) ?? orgs[0]
  } else {
    try {
      org = await createTrialOrg()
    } catch (error) {
      trackEvent({
        event: 'org_create',
        status: 'error',
        user,
        props: { source: 'onboarding', err: (error as TAPIError)?.error },
      })
      throw error
    }
    // The org's super property is not registered until OrgProvider mounts.
    trackEvent({ event: 'org_create', status: 'ok', user, props: { orgId: org.id, source: 'onboarding' } })
  }
  setOrgSession(org.id as string)
  return org
}

export interface IFirstRunResume {
  // Whether to skip the intro and open the wizard directly.
  started: boolean
  path: TPath
  cloud: TCloud
  step: TFirstRunStep
  sharedData: Record<string, unknown>
}

const IDS_TIED_TO_APP: (keyof TFirstRunMetadata)[] = [
  'app_id',
  'app_branch_id',
  'install_id',
  'workflow_id',
]

const exists = async (check: () => Promise<unknown>) => {
  try {
    await check()
    return true
  } catch (error) {
    if (statusOf(error) === 404) return false
    throw error
  }
}

export async function resolveFirstRunResume({
  orgId,
  journey,
  metadata,
  forceStart,
}: {
  orgId: string
  journey?: TUserJourney
  metadata: TFirstRunMetadata
  // Returning from GitHub or Re-open onboarding: open the wizard, not the intro.
  forceStart: boolean
}): Promise<IFirstRunResume> {
  const path: TPath = metadata.path === 'own' ? 'own' : 'example'
  const cloud: TCloud = isCloud(metadata.cloud) ? metadata.cloud : 'aws'
  const sharedData: Record<string, unknown> = {
    ...metadata,
    path,
    cloud,
    region: metadata.region ?? defaultRegion(cloud),
    testCloud: path === 'own' && isCloud(metadata.cloud) ? metadata.cloud : undefined,
  }

  const hasProgress =
    Object.keys(metadata).length > 0 || Boolean(journey?.steps?.some((step) => step.complete))
  let step: TFirstRunStep = firstIncompleteStep(journey) ?? FIRST_RUN_STEPS[0].name

  const appGone =
    !!metadata.app_id && !(await exists(() => getApp({ orgId, appId: metadata.app_id! })))
  const installGone =
    !appGone &&
    !!metadata.install_id &&
    !(await exists(() => getInstall({ orgId, installId: metadata.install_id! })))

  if (appGone || installGone) {
    IDS_TIED_TO_APP.forEach((key) => delete sharedData[key])
    step = 'start'
  }

  return { started: forceStart || hasProgress, path, cloud, step, sharedData }
}
