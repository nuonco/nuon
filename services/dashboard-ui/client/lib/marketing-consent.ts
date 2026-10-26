import { getCookie, setCookie } from '@/lib/cookies'
import { postMarketingConsent } from '@/lib/ctl-api/general'

const COOKIE = 'nuon_marketing_consent'

export async function syncMarketingConsent() {
  if (getCookie(COOKIE) !== '1') return

  try {
    const { synced } = await postMarketingConsent()
    if (!synced) return
    const host = window.location.hostname
    setCookie(COOKIE, 'synced', host.includes('.') ? host : undefined)
  } catch (error) {
    console.warn('Failed to sync marketing consent:', error)
  }
}
