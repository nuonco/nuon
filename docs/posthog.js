// PostHog analytics. Loaded automatically by Mintlify on every docs page.
// Matches the marketing site (services/website PostHog.astro in mono):
// cookieless_mode hashes visitors server-side, so no cookie banner is needed.
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
      cookieless_mode: 'always',
      disable_session_recording: true,
    })
  }
  ;(document.head || document.body).appendChild(loader)
})()
