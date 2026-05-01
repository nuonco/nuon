// Wistia player loader for popover embeds.
// Mintlify auto-includes any .js file in the content directory globally,
// so this runs once per page load and registers the <wistia-player> custom element.
(function () {
  function loadScript(src, type) {
    if (document.querySelector('script[src="' + src + '"]')) return
    var s = document.createElement('script')
    s.src = src
    s.async = true
    if (type) s.type = type
    document.head.appendChild(s)
  }

  loadScript('https://fast.wistia.com/player.js')
  loadScript('https://fast.wistia.com/embed/91y66l4t5r.js', 'module')
})()
