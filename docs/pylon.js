// Pylon chat widget. Loaded automatically by Mintlify on every docs page
// (Mintlify includes any .js file in the content directory globally).
// Same APP_ID as the dashboard-ui (server-injected as `pylonAppId` there).
(function () {
  var PYLON_APP_ID = '174f6ad2-124e-4a3b-bf7f-e80bbb2cb232'

  // Pylon's loader polls for `window.pylon.chat_settings.app_id` and won't
  // mount the iframe without it. Anonymous (no email/name) is fine.
  window.pylon = {
    chat_settings: {
      app_id: PYLON_APP_ID,
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
