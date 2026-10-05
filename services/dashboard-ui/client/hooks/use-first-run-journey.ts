import { useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  completeUserJourney,
  createUserJourney,
  getAccount,
  resetUserJourney,
  updateUserJourneyStepMetadata,
} from '@/lib'
import type { TAccount, TAPIError, TUserJourney } from '@/types'

export const FIRST_RUN_JOURNEY = 'first_run'

export const FIRST_RUN_STEPS = [
  { name: 'start', title: 'Start' },
  { name: 'connect', title: 'Connect' },
  { name: 'deploy', title: 'Deploy' },
  { name: 'stack', title: 'Stack' },
  { name: 'provision', title: 'Provision' },
] as const

export type TFirstRunStep = (typeof FIRST_RUN_STEPS)[number]['name']

export const FIRST_RUN_METADATA_KEYS = [
  'path',
  'cloud',
  'app_id',
  'app_name',
  'app_branch_id',
  'repo',
  'install_id',
  'workflow_id',
  'region',
] as const

export type TFirstRunMetadataKey = (typeof FIRST_RUN_METADATA_KEYS)[number]
export type TFirstRunMetadata = Partial<Record<TFirstRunMetadataKey, string>>

const FIRST_RUN_QUERY_KEY = ['first-run-account']

export const findFirstRun = (account?: TAccount | null): TUserJourney | undefined =>
  (account?.user_journeys as TUserJourney[] | undefined)?.find(
    (journey) => journey.name === FIRST_RUN_JOURNEY
  )

export const readFirstRunMetadata = (journey?: TUserJourney): TFirstRunMetadata => {
  const merged: TFirstRunMetadata = {}
  for (const step of journey?.steps ?? []) {
    for (const key of FIRST_RUN_METADATA_KEYS) {
      const value = step.metadata?.[key]
      if (typeof value === 'string' && value !== '') merged[key] = value
    }
  }
  return merged
}

export const firstIncompleteStep = (journey?: TUserJourney): TFirstRunStep | undefined =>
  journey?.steps?.find((step) => !step.complete)?.name as TFirstRunStep | undefined

export const isFirstRunStepComplete = (journey: TUserJourney | undefined, step: TFirstRunStep) =>
  Boolean(journey?.steps?.find((s) => s.name === step)?.complete)

async function ensureFirstRunJourney(): Promise<TAccount> {
  const account = await getAccount()
  if (findFirstRun(account)) return account
  try {
    return await createUserJourney({
      body: {
        name: FIRST_RUN_JOURNEY,
        title: 'First run',
        steps: FIRST_RUN_STEPS.map((step) => ({ ...step })),
      },
    })
  } catch (error) {
    if ((error as TAPIError)?.status === 409) return getAccount()
    console.warn('first_run journey unavailable; onboarding progress will not be saved', error)
    return account
  }
}

export async function resetFirstRunJourney(): Promise<void> {
  let account: TAccount
  try {
    account = await resetUserJourney({ journeyName: FIRST_RUN_JOURNEY })
  } catch (error) {
    if ((error as TAPIError)?.status === 404) return
    throw error
  }
  const cleared = Object.fromEntries(
    [...FIRST_RUN_METADATA_KEYS, 'skipped'].map((key) => [key, ''])
  )
  for (const step of findFirstRun(account)?.steps ?? []) {
    const hasValues = Object.values(step.metadata ?? {}).some((value) => value !== '')
    if (!hasValues) continue
    await updateUserJourneyStepMetadata({
      journeyName: FIRST_RUN_JOURNEY,
      stepName: step.name,
      metadata: cleared,
      complete: false,
    })
  }
}

export interface IFirstRunJourney {
  isReady: boolean
  isLoading: boolean
  error: TAPIError | null
  journey: TUserJourney | undefined
  metadata: TFirstRunMetadata
  saveStep: (
    step: TFirstRunStep,
    metadata: Record<string, string>,
    opts?: { complete?: boolean }
  ) => Promise<void>
  complete: () => Promise<void>
  skip: (step: TFirstRunStep) => Promise<void>
}

export function useFirstRunJourney({ enabled = true }: { enabled?: boolean } = {}): IFirstRunJourney {
  const queryClient = useQueryClient()
  const { data: account, isLoading, error } = useQuery<TAccount, TAPIError>({
    queryKey: FIRST_RUN_QUERY_KEY,
    queryFn: ensureFirstRunJourney,
    enabled,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  })

  const journey = findFirstRun(account)

  const saveStep = useCallback<IFirstRunJourney['saveStep']>(
    async (step, metadata, opts = {}) => {
      const current = findFirstRun(queryClient.getQueryData<TAccount>(FIRST_RUN_QUERY_KEY))
      if (!current) return
      const complete = opts.complete ?? isFirstRunStepComplete(current, step)
      const updated = await updateUserJourneyStepMetadata({
        journeyName: FIRST_RUN_JOURNEY,
        stepName: step,
        metadata,
        complete,
      })
      queryClient.setQueryData(FIRST_RUN_QUERY_KEY, updated)
    },
    [queryClient]
  )

  const complete = useCallback(async () => {
    if (!findFirstRun(queryClient.getQueryData<TAccount>(FIRST_RUN_QUERY_KEY))) return
    const updated = await completeUserJourney({ journeyName: FIRST_RUN_JOURNEY })
    queryClient.setQueryData(FIRST_RUN_QUERY_KEY, updated)
  }, [queryClient])

  const skip = useCallback(
    async (step: TFirstRunStep) => {
      await saveStep(step, { skipped: 'true' })
      await complete()
    },
    [saveStep, complete]
  )

  return {
    isReady: !!account,
    isLoading,
    error: error ?? null,
    journey,
    metadata: readFirstRunMetadata(journey),
    saveStep,
    complete,
    skip,
  }
}
