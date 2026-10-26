import { getMarketingConsent, setMarketingConsentSynced } from '@/lib/cookies'
import { postMarketingConsent } from '@/lib/ctl-api/general'

export async function syncMarketingConsent() {
  if (getMarketingConsent() !== '1') return

  try {
    const { synced } = await postMarketingConsent()
    if (synced) setMarketingConsentSynced()
  } catch (error) {
    console.warn('Failed to sync marketing consent:', error)
  }
}
