import { api } from '@/lib/api'
import type { TAPIVersion } from '@/types'

export async function getAPIVersion() {
  return api<TAPIVersion>({
    pathVersion: '',
    path: 'version',
  })
}
