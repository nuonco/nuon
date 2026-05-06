import { api } from '../api'

export const triggerPromotion = () =>
  api<{ status: string; tag: string }>({ path: 'promote', method: 'POST' })

export const triggerSeed = () =>
  api<string>({ path: 'seed', method: 'POST' })
