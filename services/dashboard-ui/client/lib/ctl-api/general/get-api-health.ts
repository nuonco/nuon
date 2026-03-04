import { api } from '@/lib/api'
import type { TAPIHealth } from '@/types'

export async function getAPIHealth() {
  return api<TAPIHealth>({
    pathVersion: '',
    path: 'livez',
  })
}
