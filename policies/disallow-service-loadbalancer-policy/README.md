[![Kubewarden Policy Repository](https://github.com/kubewarden/community/blob/main/badges/kubewarden-policies.svg)](https://github.com/kubewarden/community/blob/main/REPOSITORIES.md#policy-scope)
[![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/kubewarden/community/blob/main/REPOSITORIES.md#stable)

This policy works by inspecting `type` of `Service` resources and prevents the
creation of Services with type `LoadBalancer`.

Kubernetes network policies have no control over what is being exposed
via these type of Services. Moreover, on public clouds, the creation of
`LoadBalancer` Services leads to additional charges.

Because of that, it's usually a safer choice to have
tighter control over the creation of these type of Services.

# Configuration

This policy doesn't take any configuration value.
