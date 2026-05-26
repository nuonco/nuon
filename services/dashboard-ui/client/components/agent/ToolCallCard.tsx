import type { TAgentToolCall } from '@/providers/agent-provider'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

const toolLabels: Record<string, string> = {
  list_apps: 'Listing apps',
  get_app: 'Getting app details',
  create_app: 'Creating app',
  get_app_config: 'Getting app config',
  list_components: 'Listing components',
  create_component: 'Creating component',
  create_terraform_config: 'Configuring Terraform module',
  create_helm_config: 'Configuring Helm chart',
  create_k8s_manifest_config: 'Configuring K8s manifest',
  create_docker_build_config: 'Configuring Docker build',
  build_all_components: 'Building all components',
  get_build: 'Checking build status',
  list_vcs_connections: 'Listing VCS connections',
  list_repos: 'Listing repositories',
  list_branches: 'Listing branches',
  list_installs: 'Listing installs',
  get_install: 'Getting install details',
  create_install: 'Creating install',
  get_install_inputs: 'Getting install inputs',
  get_cloud_regions: 'Getting cloud regions',
  list_deploys: 'Listing deploys',
  get_workflows: 'Getting workflows',
  get_workflow_steps: 'Getting workflow steps',
  get_step_logs: 'Reading logs',
  get_runner_job_plan: 'Reading Terraform plan',
  get_runner: 'Getting runner details',
  get_runner_health: 'Checking runner health',
  run_adhoc_action: 'Running action',
}

const statusMap: Record<string, string> = {
  running: 'executing',
  complete: 'success',
  error: 'failed',
}

interface IToolCallCard {
  toolCall: TAgentToolCall
}

export function ToolCallCard({ toolCall }: IToolCallCard) {
  const label = toolLabels[toolCall.tool] ?? toolCall.tool
  const status = statusMap[toolCall.status] ?? 'executing'

  return (
    <div className="flex items-center gap-3">
      <Status status={status} variant="timeline" />
      <Text variant="subtext" family="mono">{label}</Text>
    </div>
  )
}
