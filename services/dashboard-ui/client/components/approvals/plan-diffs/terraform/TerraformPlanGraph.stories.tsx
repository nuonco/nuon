export default {
  title: 'Approvals/PlanDiffs/TerraformPlanGraph',
}

import { parseTerraformPlan } from '@/utils/terraform-utils'
import { TerraformPlanGraph } from './TerraformPlanGraph'

const parse = (plan: any) => {
  const { resources, drift, outputs } = parseTerraformPlan(plan)
  return { resources: resources.changes, drift: drift.changes, outputs: outputs.changes }
}

export const SimpleCreates = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'aws_eks_cluster.main',
        type: 'aws_eks_cluster',
        name: 'main',
        module_address: 'module.eks',
        change: { actions: ['create'], before: null, after: { name: 'prod-cluster', version: '1.29' } },
      },
      {
        address: 'aws_eks_node_group.default',
        type: 'aws_eks_node_group',
        name: 'default',
        module_address: 'module.eks',
        change: { actions: ['create'], before: null, after: { cluster_name: 'prod-cluster', node_group_name: 'default' } },
      },
      {
        address: 'aws_eks_addon.vpc_cni',
        type: 'aws_eks_addon',
        name: 'vpc_cni',
        module_address: 'module.eks',
        change: { actions: ['create'], before: null, after: { addon_name: 'vpc-cni' } },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const MixedActions = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'aws_vpc.main',
        type: 'aws_vpc',
        name: 'main',
        change: { actions: ['no-op'], before: { cidr_block: '10.0.0.0/16' }, after: { cidr_block: '10.0.0.0/16' } },
      },
      {
        address: 'aws_subnet.public',
        type: 'aws_subnet',
        name: 'public',
        change: { actions: ['update'], before: { map_public_ip_on_launch: false }, after: { map_public_ip_on_launch: true } },
      },
      {
        address: 'aws_route_table.private',
        type: 'aws_route_table',
        name: 'private',
        change: { actions: ['create'], before: null, after: { vpc_id: 'vpc-abc123' } },
      },
      {
        address: 'aws_security_group_rule.old_ssh',
        type: 'aws_security_group_rule',
        name: 'old_ssh',
        change: { actions: ['delete'], before: { from_port: 22 }, after: null },
      },
      {
        address: 'aws_nat_gateway.main',
        type: 'aws_nat_gateway',
        name: 'main',
        change: { actions: ['no-op'], before: { subnet_id: 'subnet-pub001' }, after: { subnet_id: 'subnet-pub001' } },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const ReplaceResources = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'aws_instance.web',
        type: 'aws_instance',
        name: 'web',
        module_address: 'module.compute',
        change: {
          actions: ['delete', 'create'],
          before: { ami: 'ami-old', instance_type: 't3.medium' },
          after: { ami: 'ami-new', instance_type: 't3.medium' },
        },
      },
      {
        address: 'aws_eip.web',
        type: 'aws_eip',
        name: 'web',
        module_address: 'module.compute',
        change: {
          actions: ['create', 'delete'],
          before: { public_ip: '54.200.100.50' },
          after: { public_ip: 'Known after apply' },
        },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const WithModules = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'aws_iam_role.ecs_task_execution',
        type: 'aws_iam_role',
        name: 'ecs_task_execution',
        module_address: 'module.ecs',
        change: {
          actions: ['update'],
          before: { name: 'ecs-task-execution', max_session_duration: 3600 },
          after: { name: 'ecs-task-execution', max_session_duration: 7200 },
        },
      },
      {
        address: 'aws_iam_role_policy.ecs_secrets',
        type: 'aws_iam_role_policy',
        name: 'ecs_secrets',
        module_address: 'module.ecs',
        change: { actions: ['create'], before: null, after: { name: 'ecs-secrets-access' } },
      },
      {
        address: 'aws_s3_bucket.logs',
        type: 'aws_s3_bucket',
        name: 'logs',
        change: { actions: ['create'], before: null, after: { bucket: 'my-logs' } },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const LargePlan = () => {
  const resourceChanges: any[] = []
  const modules = ['module.network', 'module.compute', 'module.storage']
  const types = ['aws_instance', 'aws_security_group', 'aws_subnet', 'aws_s3_bucket', 'aws_iam_role', 'aws_lambda_function', 'aws_sqs_queue']
  const actions = ['create', 'update', 'delete', 'no-op']

  let idx = 0
  for (const mod of modules) {
    for (const type of types) {
      const action = actions[idx % actions.length]
      resourceChanges.push({
        address: `${mod}.${type}.item_${idx}`,
        type,
        name: `item_${idx}`,
        module_address: mod,
        change: {
          actions: [action],
          before: action === 'create' ? null : { name: `item_${idx}` },
          after: action === 'delete' ? null : { name: `item_${idx}_v2` },
        },
      })
      idx++
    }
  }

  const data = parse({ resource_changes: resourceChanges, output_changes: {} })
  return <TerraformPlanGraph {...data} />
}

export const NoOpAndRead = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'data.aws_caller_identity.current',
        type: 'aws_caller_identity',
        name: 'current',
        change: { actions: ['read'], before: null, after: { account_id: 'Known after apply' } },
      },
      {
        address: 'data.aws_region.current',
        type: 'aws_region',
        name: 'current',
        change: { actions: ['read'], before: null, after: { name: 'Known after apply' } },
      },
      {
        address: 'aws_s3_bucket.logs',
        type: 'aws_s3_bucket',
        name: 'logs',
        change: { actions: ['no-op'], before: { bucket: 'my-logs' }, after: { bucket: 'my-logs' } },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const WithDrift = () => {
  const data = parse({
    resource_drift: [
      {
        address: 'aws_autoscaling_group.web',
        type: 'aws_autoscaling_group',
        name: 'web',
        change: {
          actions: ['update'],
          before: { desired_capacity: 3 },
          after: { desired_capacity: 5 },
        },
      },
    ],
    resource_changes: [
      {
        address: 'aws_autoscaling_group.web',
        type: 'aws_autoscaling_group',
        name: 'web',
        change: {
          actions: ['update'],
          before: { desired_capacity: 5 },
          after: { desired_capacity: 3 },
        },
      },
      {
        address: 'aws_launch_template.web',
        type: 'aws_launch_template',
        name: 'web',
        change: {
          actions: ['update'],
          before: { image_id: 'ami-old' },
          after: { image_id: 'ami-new' },
        },
      },
    ],
    output_changes: {},
  })
  return <TerraformPlanGraph {...data} />
}

export const WithOutputs = () => {
  const data = parse({
    resource_changes: [
      {
        address: 'aws_db_instance.main',
        type: 'aws_db_instance',
        name: 'main',
        module_address: 'module.database',
        change: {
          actions: ['delete', 'create'],
          before: { identifier: 'myapp-db', engine_version: '14.10' },
          after: { identifier: 'myapp-db-v2', engine_version: '16.2' },
        },
      },
    ],
    output_changes: {
      db_endpoint: {
        actions: ['update'],
        before: 'myapp-db.abc123.us-west-2.rds.amazonaws.com:5432',
        after: 'Known after apply',
        after_unknown: true,
      },
      db_identifier: {
        actions: ['update'],
        before: 'myapp-db',
        after: 'myapp-db-v2',
      },
    },
  })
  return <TerraformPlanGraph {...data} />
}

export const FullPlan = () => {
  const data = parse({
    resource_drift: [
      {
        address: 'azurerm_container_registry.acr',
        type: 'azurerm_container_registry',
        name: 'acr',
        change: {
          actions: ['update'],
          before: { name: 'myregistry', tags: null },
          after: { name: 'myregistry', tags: {} },
        },
      },
      {
        address: 'module.aks.azurerm_kubernetes_cluster.main',
        module_address: 'module.aks',
        type: 'azurerm_kubernetes_cluster',
        name: 'main',
        change: {
          actions: ['update'],
          before: { name: 'my-aks', tags: null },
          after: { name: 'my-aks', tags: {} },
        },
      },
    ],
    resource_changes: [
      {
        address: 'azurerm_container_registry.acr',
        type: 'azurerm_container_registry',
        name: 'acr',
        change: { actions: ['no-op'], before: { name: 'myregistry' }, after: { name: 'myregistry' } },
      },
      {
        address: 'module.aks.azurerm_kubernetes_cluster.main',
        module_address: 'module.aks',
        type: 'azurerm_kubernetes_cluster',
        name: 'main',
        change: { actions: ['no-op'], before: { name: 'my-aks' }, after: { name: 'my-aks' } },
      },
      {
        address: 'aws_s3_bucket.new_logs',
        type: 'aws_s3_bucket',
        name: 'new_logs',
        change: { actions: ['create'], before: null, after: { bucket: 'new-logs-bucket' } },
      },
    ],
    output_changes: {
      acr_name: {
        actions: ['no-op'],
        before: 'myregistry',
        after: 'myregistry',
      },
      cluster_name: {
        actions: ['no-op'],
        before: 'my-aks',
        after: 'my-aks',
      },
    },
  })
  return <TerraformPlanGraph {...data} />
}

export const Empty = () => (
  <TerraformPlanGraph resources={[]} drift={[]} outputs={[]} />
)
