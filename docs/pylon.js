// Pylon chat widget. Loaded automatically by Mintlify on every docs page
// (Mintlify includes any .js file in the content directory globally).
// Same APP_ID as the dashboard-ui (server-injected as `pylonAppId` there).
(function () {
  var PYLON_APP_ID = '174f6ad2-124e-4a3b-bf7f-e80bbb2cb232'

  // Pylon's bundle silently refuses to mount unless both email and name are
  // set on chat_settings (validation errors are swallowed by their internal
  // telemetry sink — no console signal). For unauthenticated docs visitors
  // we generate a stable per-browser identity so a returning reader resumes
  // their thread instead of opening a fresh contact each visit.
  var VISITOR_KEY = 'nuon-docs-pylon-visitor-id'
  var visitorId
  try {
    visitorId = localStorage.getItem(VISITOR_KEY)
    if (!visitorId) {
      visitorId =
        (window.crypto && crypto.randomUUID && crypto.randomUUID()) ||
        Date.now().toString(36) + Math.random().toString(36).slice(2)
      localStorage.setItem(VISITOR_KEY, visitorId)
    }
  } catch (_) {
    visitorId = Date.now().toString(36) + Math.random().toString(36).slice(2)
  }

  window.pylon = {
    chat_settings: {
      app_id: PYLON_APP_ID,
      email: 'docs-visitor-' + visitorId + '@docs.nuon.co',
      name: 'Docs Visitor',
    },
  }

  var e = window
  var t = document
  var n = function () {
    n.e(arguments)
  }
  n.q = []
  n.e = function (e) {
    n.q.push(e)
  }
  e.Pylon = n

  var r = function () {
    var e = t.createElement('script')
    e.setAttribute('type', 'text/javascript')
    e.setAttribute('async', 'true')
    e.setAttribute('src', 'https://widget.usepylon.com/widget/' + PYLON_APP_ID)
    var n = t.getElementsByTagName('script')[0]
    if (n && n.parentNode) {
      n.parentNode.insertBefore(e, n)
    } else {
      t.head.appendChild(e)
    }
  }

  if (t.readyState === 'complete') {
    r()
  } else {
    e.addEventListener('load', r, false)
  }
})()
