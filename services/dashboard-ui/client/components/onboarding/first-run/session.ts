import { FIRST_RUN_STEPS, type TFirstRunStep } from '@/hooks/use-first-run-journey'
import { isCloud, type TCloud, type TPath } from './constants'

const STORAGE_PREFIX = 'nuon.first-run-onboarding'

const STEP_NAMES = new Set<string>(FIRST_RUN_STEPS.map((step) => step.name))

export interface IFirstRunSession {
  started: boolean
  step: TFirstRunStep
  path: TPath
  cloud: TCloud
  sharedData: Record<string, unknown>
}

export const firstRunStorageKey = (userId?: string) =>
  userId ? `${STORAGE_PREFIX}:${userId}` : STORAGE_PREFIX

const isStep = (value: unknown): value is TFirstRunStep =>
  typeof value === 'string' && STEP_NAMES.has(value)

const isPath = (value: unknown): value is TPath => value === 'own' || value === 'example'

export function readFirstRunSession(
  userId?: string,
  storage: Pick<Storage, 'getItem'> = localStorage
): IFirstRunSession | undefined {
  try {
    const raw = storage.getItem(firstRunStorageKey(userId))
    if (!raw) return undefined
    const parsed = JSON.parse(raw) as Partial<IFirstRunSession>
    if (!parsed.started || !isStep(parsed.step) || !isPath(parsed.path) || !isCloud(parsed.cloud)) {
      return undefined
    }
    const sharedData =
      parsed.sharedData && typeof parsed.sharedData === 'object' ? parsed.sharedData : {}
    return {
      started: true,
      step: parsed.step,
      path: parsed.path,
      cloud: parsed.cloud,
      sharedData,
    }
  } catch {
    return undefined
  }
}

export function writeFirstRunSession(
  userId: string | undefined,
  session: IFirstRunSession,
  storage: Pick<Storage, 'setItem'> = localStorage
) {
  try {
    storage.setItem(firstRunStorageKey(userId), JSON.stringify(session))
  } catch {}
}

export function clearFirstRunSession(
  userId?: string,
  storage: Pick<Storage, 'removeItem'> = localStorage
) {
  try {
    storage.removeItem(firstRunStorageKey(userId))
  } catch {}
}
