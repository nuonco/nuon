// PostHog analytics. Loaded automatically by Mintlify on every docs page.
// Direct US ingest (us.i.posthog.com) — not Mintlify's ph.mintlify.com proxy.
// Do not set cookieless_mode: 'always': that strips $ip before GeoIP runs, so
// country/city enrichment never lands (see PostHog cookieless docs). Prefer
// project "Discard client IP data" if you need to drop raw IPs after enrichment.
(function () {
  if (window.__posthogInitialized) return
  window.__posthogInitialized = true

  var POSTHOG_KEY = 'phc_1NEQAphH0jxCX7opmp4Iyq2O6mu4tM552kMjJz8uKkl'
  var POSTHOG_HOST = 'https://us.i.posthog.com'

  var loader = document.createElement('script')
  loader.async = true
  loader.crossOrigin = 'anonymous'
  loader.src = 'https://us-assets.i.posthog.com/static/array.js'
  loader.onload = function () {
    if (!window.posthog || typeof window.posthog.init !== 'function') return
    window.posthog.init(POSTHOG_KEY, {
      api_host: POSTHOG_HOST,
      autocapture: true,
      capture_pageview: 'history_change',
      person_profiles: 'never',
      disable_session_recording: true,
    })
  }
  ;(document.head || document.body).appendChild(loader)
})()
