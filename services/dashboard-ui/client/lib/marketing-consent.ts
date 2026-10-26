// The sign-in page sets this cookie to 1 when the consent box is ticked. Once the user is added
// to the Mailchimp audience it is set to "synced", which keeps the box hidden on later sign-ins.
const COOKIE = 'nuon_marketing_consent'

function writeCookie(value: string) {
  const host = window.location.hostname
  document.cookie =
    `${COOKIE}=${value}; path=/; max-age=31536000; SameSite=Lax` +
    (host.includes('.') ? `; domain=${host}` : '') +
    (window.location.protocol === 'https:' ? '; Secure' : '')
}

export async function syncMarketingConsent() {
  if (!document.cookie.split(';').some((c) => c.trim() === `${COOKIE}=1`))
    return

  try {
    const res = await fetch('/api/marketing-consent', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) return
    const { synced } = (await res.json()) as { synced: boolean }
    if (synced) writeCookie('synced')
  } catch (error) {
    console.warn('Failed to sync marketing consent:', error)
  }
}
