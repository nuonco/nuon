import { api } from '@/lib/api'
import type { TCLIConfig } from '@/types'

export async function getCLIConfig() {
  return api<TCLIConfig>({
    path: 'general/cli-config',
  })
}
