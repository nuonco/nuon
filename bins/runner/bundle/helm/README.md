# Runner Helm Chart

## Examples

**One runner with its own node group:**

`node_pool.enabled=true` is the default in `values.yaml`.

```sh
helm install rug2klm0zqhap3lstkffiqancbsr3w . \
	--set nuon_api_token="tok2kLlzVm7K6Gicxb0sGwfcGbGQaT",runner_group_id="rug2kLm0ZQHAP3lsTKFfiQaNCbsR3W" \
	--namespace rug2klm0zqhap3lstkffiqancbsr3w \
	--create-namespace --dry-run
```

**Multiple runners and its own node group:**

`node_pool.enabled=true` is the default in `values.yaml`.

Note: If we want a deployment with more than one runner, it SHOULD have its own
node group.

```sh
helm install rug2klm0zqhap3lstkffiqancbsr3w . \
	--set nuon_api_token="tok2kLlzVm7K6Gicxb0sGwfcGbGQaT",runner_group_id="rug2kLm0ZQHAP3lsTKFfiQaNCbsR3W",node_pool.runner_count="5" \
	--namespace rug2klm0zqhap3lstkffiqancbsr3w \
	--create-namespace --dry-run
```

## Notes

1. In practice, the `Namespace` and `ReleaseName` are both the lower-case
   `runner_group_id`.
2. The `runner_group_id` is used to label all of the resources.

### NodePool Notes

1. The `NodePool` for the runner has taints set on it. The matching deployment
   has matching tolerations. This creates a robust resource that gives us a
   stronger guarantee around resource availability.
2. At the time of writing, the `NodePool`s all use the same `EC2NodeClass`
   (default). We can support different instance types but that requires the
   creation of additional `EC2NodeClass`es. This helm chart is likely a good
   place to do it, as opposed to doing it in `infra/eks`.
3. The `runner_count` and `instance_type` determines the resource limits on the
   `NodePool`. We set the limits to an amount equivalen to `runner_count` nodes
   plus one node. See the limits section in `./node_pool.tpl` for detail.
   - `memory` is set as an int in `values.yaml` but treated as `Mi`.
   - `cpu` is simple.
