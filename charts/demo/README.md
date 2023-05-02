# Nuon demo chart

This Helm chart is useful for demoing Nuon's support for deploying Helm charts.

## Resources

### Configmap

We provision a configmap, that accepts the values of `env` as inputs. This configmap is designed to be mounted for environment variables in a running deployment.

