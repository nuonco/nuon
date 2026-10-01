import { expect, test } from 'bun:test'
import type { TUserJourney } from '@/types'
import { resolveFirstRunResume } from './resume'
import {
  firstRunStorageKey,
  readFirstRunSession,
  writeFirstRunSession,
  type IFirstRunSession,
} from './session'

function memoryStorage(): Storage {
  const data = new Map<string, string>()
  return {
    get length() {
      return data.size
    },
    clear() {
      data.clear()
    },
    getItem(key) {
      return data.get(key) ?? null
    },
    key(index) {
      return [...data.keys()][index] ?? null
    },
    removeItem(key) {
      data.delete(key)
    },
    setItem(key, value) {
      data.set(key, value)
    },
  }
}

const journeyOnConnect = {
  name: 'first_run',
  title: 'First run',
  steps: [
    {
      name: 'start',
      title: 'Start',
      complete: true,
      completed_at: null,
      completion_method: null,
      completion_source: null,
      metadata: {},
    },
    {
      name: 'connect',
      title: 'Connect',
      complete: false,
      completed_at: null,
      completion_method: null,
      completion_source: null,
      metadata: {},
    },
  ],
} satisfies TUserJourney

test('the stored step survives a reload for that user', () => {
  const storage = memoryStorage()
  const session: IFirstRunSession = {
    started: true,
    step: 'stack',
    path: 'example',
    cloud: 'aws',
    sharedData: { app_id: 'app-1', region: 'us-west-2' },
  }
  writeFirstRunSession('user-1', session, storage)
  expect(readFirstRunSession('user-1', storage)).toEqual(session)
  expect(readFirstRunSession('user-2', storage)).toBeUndefined()
})

test('a broken stored session is ignored', () => {
  const storage = memoryStorage()
  storage.setItem(firstRunStorageKey('user-1'), '{')
  expect(readFirstRunSession('user-1', storage)).toBeUndefined()
})

test('resume opens the stored step when the journey is somewhere else', async () => {
  const resume = await resolveFirstRunResume({
    orgId: 'org-1',
    journey: journeyOnConnect,
    metadata: { path: 'own', cloud: 'aws' },
    forceStart: false,
    session: {
      started: true,
      step: 'stack',
      path: 'example',
      cloud: 'gcp',
      sharedData: { app_name: 'acme', region: 'us-central1' },
    },
  })

  expect(resume.started).toBe(true)
  expect(resume.step).toBe('stack')
  expect(resume.path).toBe('example')
  expect(resume.cloud).toBe('gcp')
  expect(resume.sharedData.app_name).toBe('acme')
  expect(resume.sharedData.region).toBe('us-central1')
})
