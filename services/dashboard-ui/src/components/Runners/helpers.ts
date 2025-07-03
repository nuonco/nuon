import type { TRunnerJob } from '@/types'

export function jobHrefPath(job: TRunnerJob): string {
  let hrefPath: string

  switch (job?.group) {
    case 'build':
      hrefPath = `apps/${job?.metadata?.app_id}/components/${job?.metadata?.component_id}/builds/${job?.metadata?.component_build_id}`
      break
    case 'sandbox':
      hrefPath = `installs/${job?.metadata?.install_id}/sandbox/${job?.metadata?.sandbox_run_id}`
      break
    case 'sync':
      hrefPath = `installs/${job?.metadata?.install_id}/components/${job?.metadata?.component_id}/deploys/${job?.metadata?.deploy_id}`
      break
    case 'deploy':
      hrefPath = `installs/${job?.metadata?.install_id}/components/${job?.metadata?.component_id}/deploys/${job?.metadata?.deploy_id}`
      break
    case 'actions':
      hrefPath = `installs/${job?.metadata?.install_id}/actions/${job?.metadata?.action_workflow_id}/${job?.metadata?.action_workflow_run_id}`
      break
    default:
      hrefPath = ''
  }

  return hrefPath
}

export function jobName(job: TRunnerJob): string {
  let name: string

  switch (job?.group) {
    case 'build':
      name = job?.metadata?.component_name
      break
    case 'sandbox':
      name = job?.metadata?.sandbox_run_type
      break
    case 'sync':
      name = job?.metadata?.component_name
      break
    case 'deploy':
      name = job?.metadata?.component_name
      break
    case 'actions':
      name = job?.metadata?.action_workflow_name
      break
    case 'operations':
      name = job?.type
      break
    default:
      name = 'Unknown'
  }

  return name
}

export function jobOperation(job: TRunnerJob): string {
  let operation: string

  switch (job?.operation) {
    case 'create-apply-plan':
      operation = 'plan'
      break
    case 'create-teardown-plan':
      operation = 'teardown plan'
      break
    case 'apply-plan':
      operation = 'apply'
      break
    case 'exec':
      operation = 'run'
      break
    case 'build':
      operation = job?.group === 'sync' ? 'sync' : 'build'
      break
    default:
      operation = ''
  }

  return operation
}
