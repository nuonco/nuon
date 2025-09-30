// REB2B Integration
(function (key) {
  if (window.reb2b) return;

  window.reb2b = { loaded: true };

  var script = document.createElement("script");
  script.async = true;
  script.src =
    "https://ddwl4m2hdecbv.cloudfront.net/b/" + key + "/" + key + ".js.gz";

  var firstScript = document.getElementsByTagName("script")[0];
  firstScript.parentNode.insertBefore(script, firstScript);
})("W6Z57HQ48XOX");

// Unify Intent Integration
(function () {
  var methods = [
    "identify",
    "page",
    "startAutoPage",
    "stopAutoPage",
    "startAutoIdentify",
    "stopAutoIdentify",
  ];

  function createProxy(queue) {
    return Object.assign(
      [],
      methods.reduce(function (proxy, method) {
        proxy[method] = function () {
          queue.push([method, [].slice.call(arguments)]);
          return queue;
        };
        return proxy;
      }, {})
    );
  }

  window.unify = window.unify || createProxy(window.unify || []);
  window.unifyBrowser =
    window.unifyBrowser || createProxy(window.unifyBrowser || []);

  var unifyScript = document.createElement("script");
  unifyScript.async = true;
  unifyScript.setAttribute(
    "src",
    "https://tag.unifyintent.com/v1/TzHRecTDNzeGwGWox2kSzp/script.js"
  );
  unifyScript.setAttribute(
    "data-api-key",
    "wk_NoCsMGud_AP3KdrDCFsen1xC8Uup1YWd1fkRX8Yzj"
  );
  unifyScript.setAttribute("id", "unifytag");

  (document.body || document.head).appendChild(unifyScript);
})();
