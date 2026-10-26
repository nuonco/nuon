import { api } from '@/lib/api'

export async function postMarketingConsent() {
  return api<{ synced: boolean }>({
    method: 'POST',
    pathVersion: '',
    path: 'api/marketing-consent',
  })
}
