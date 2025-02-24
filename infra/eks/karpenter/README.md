# Karpenter Sub Module

## Additional Nodepools

### Taints

As written, we do not support additional or custom taints. Each nodepool gets a
taint automatically.

```hcl
{
    key    = "pool.nuon.co"
    value  = each.value.name
    effect = "NoSchedule"
}
```
