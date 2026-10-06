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
import type { IFirstRunSession } from './session'

const statusOf = (error: unknown) => (error as TAPIError | undefined)?.status

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
    trackEvent({ event: 'org_create', status: 'ok', user, props: { orgId: org.id, source: 'onboarding' } })
  }
  setOrgSession(org.id as string)
  return org
}

export interface IFirstRunResume {
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

const readId = (value: unknown) => (typeof value === 'string' && value !== '' ? value : undefined)

export async function resolveFirstRunResume({
  orgId,
  journey,
  metadata,
  forceStart,
  session,
}: {
  orgId: string
  journey?: TUserJourney
  metadata: TFirstRunMetadata
  forceStart: boolean
  session?: IFirstRunSession
}): Promise<IFirstRunResume> {
  const path: TPath = session?.path ?? (metadata.path === 'own' ? 'own' : 'example')
  const cloud: TCloud = session?.cloud ?? (isCloud(metadata.cloud) ? metadata.cloud : 'aws')
  const sharedData: Record<string, unknown> = {
    ...metadata,
    ...session?.sharedData,
    path,
    cloud,
    region: readId(session?.sharedData.region) ?? metadata.region ?? defaultRegion(cloud),
    testCloud: path === 'own' && isCloud(cloud) ? cloud : undefined,
  }

  const hasProgress =
    !!session?.started ||
    Object.keys(metadata).length > 0 ||
    Boolean(journey?.steps?.some((step) => step.complete))
  let step: TFirstRunStep = session?.step ?? firstIncompleteStep(journey) ?? FIRST_RUN_STEPS[0].name

  const appId = readId(sharedData.app_id)
  const installId = readId(sharedData.install_id)
  const appGone = !!appId && !(await exists(() => getApp({ orgId, appId })))
  const installGone = !appGone && !!installId && !(await exists(() => getInstall({ orgId, installId })))

  if (appGone || installGone) {
    IDS_TIED_TO_APP.forEach((key) => delete sharedData[key])
    step = 'start'
  }

  return { started: forceStart || hasProgress, path, cloud, step, sharedData }
}
