Return the allowed hosts that apply to an app's config.

The control plane operator can restrict where URLs in an app config may point, such as the runner API URL, the runner init script, and stack template URLs. Each category is either unrestricted or carries a list of allowed URL prefixes. A sync that sets a URL outside a restricted category is rejected, or only reported when the mode is `warn`.
