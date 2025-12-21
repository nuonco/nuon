import type { TWorkflow } from '@/types'

// Status types for workflow stages and steps
export type TWorkflowStageStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

// Step interface
export interface IWorkflowStep {
  id: string
  name: string
  status: TWorkflowStageStatus
  executionTime?: number
  startedAt?: string
  completedAt?: string
  message?: string
  error?: string
  substeps?: IWorkflowStep[]
  logs?: string[]
}

// Parallel install/component update interface
export interface IParallelInstallUpdate {
  id: string
  installName: string
  status: TWorkflowStageStatus
  startedAt?: string
  completedAt?: string
  executionTime?: number
  steps: IWorkflowStep[]
}

// Stage interface
export interface IWorkflowStage {
  id: string
  name: string
  description: string
  status: TWorkflowStageStatus
  startedAt?: string
  completedAt?: string
  executionTime?: number
  metadata?: Record<string, any>
  steps: IWorkflowStep[]
  parallelInstalls?: IParallelInstallUpdate[]
}

function mapStepStatus(step: any): TWorkflowStageStatus {
  if (!step.status) return 'pending'

  const status = step.status.toLowerCase()

  if (status.includes('complete') || status.includes('success')) {
    return 'completed'
  }
  if (status.includes('fail') || status.includes('error')) {
    return 'failed'
  }
  if (status.includes('running') || status.includes('in-progress') || status.includes('in_progress')) {
    return 'running'
  }
  if (status.includes('cancel')) {
    return 'cancelled'
  }

  if (step.finished === true) {
    return 'completed'
  }
  if (step.finished === false && step.started_at) {
    return 'running'
  }

  return 'pending'
}

function determineStageStatus(steps: IWorkflowStep[]): TWorkflowStageStatus {
  if (steps.length === 0) return 'pending'

  const hasRunning = steps.some(s => s.status === 'running')
  const hasFailed = steps.some(s => s.status === 'failed')
  const allCompleted = steps.every(s => s.status === 'completed')

  if (hasFailed) return 'failed'
  if (hasRunning) return 'running'
  if (allCompleted) return 'completed'

  return 'pending'
}

function generateStepName(step: any): string {
  if (step.name) return step.name

  const targetType = step.step_target_type || ''
  
  if (targetType.includes('deploy')) {
    return 'Deploy Component'
  }
  if (targetType.includes('build')) {
    return 'Build Component'
  }
  if (targetType.includes('sandbox')) {
    return 'Run Sandbox'
  }
  if (targetType.includes('action')) {
    return 'Execute Action'
  }
  if (targetType.includes('approval')) {
    return 'Pending Approval'
  }

  return `Step ${step.idx || 'Unknown'}`
}

function generateStageName(groupIdx: number, steps: any[]): string {
  if (steps.length === 0) return `Stage ${groupIdx + 1}`

  const hasBuilds = steps.some(s => s.step_target_type?.includes('build'))
  const hasDeploys = steps.some(s => s.step_target_type?.includes('deploy'))
  const hasApprovals = steps.some(s => s.step_target_type?.includes('approval'))
  const hasSandbox = steps.some(s => s.step_target_type?.includes('sandbox'))

  if (hasApprovals) {
    return 'Pending Approval'
  }
  if (hasBuilds) {
    return steps.length > 1 ? 'Build Components' : 'Build Component'
  }
  if (hasDeploys) {
    return steps.length > 1 ? 'Deploy to Installs' : 'Deploy to Install'
  }
  if (hasSandbox) {
    return 'Run Sandbox'
  }

  return `Stage ${groupIdx + 1}`
}

function transformWorkflowStep(step: any): IWorkflowStep {
  const status = mapStepStatus(step)
  const name = generateStepName(step)

  return {
    id: step.id || `step-${step.idx}`,
    name,
    status,
    executionTime: step.execution_time,
    startedAt: step.created_at,
    completedAt: step.finished_at,
    message: step.status_description,
    error: status === 'failed' ? step.status_description : undefined,
  }
}

export function transformWorkflowToStages(workflow: TWorkflow): IWorkflowStage[] {
  if (!workflow.steps || workflow.steps.length === 0) {
    return []
  }

  const stepsByGroup = new Map<number, any[]>()
  
  workflow.steps.forEach(step => {
    const groupIdx = step.group_idx ?? 0
    const existing = stepsByGroup.get(groupIdx) || []
    existing.push(step)
    stepsByGroup.set(groupIdx, existing)
  })

  const sortedGroups = Array.from(stepsByGroup.entries()).sort((a, b) => a[0] - b[0])

  const stages: IWorkflowStage[] = sortedGroups.map(([groupIdx, groupSteps]) => {
    const sortedSteps = groupSteps.sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0))

    const isParallelStage = sortedSteps.length > 1

    const stageName = generateStageName(groupIdx, sortedSteps)
    const stageId = `stage-${groupIdx}`

    let steps: IWorkflowStep[] = []
    let parallelInstalls: IParallelInstallUpdate[] | undefined

    if (isParallelStage) {
      parallelInstalls = sortedSteps.map((step, idx) => {
        const transformedStep = transformWorkflowStep(step)
        const installName = step.name || `Operation ${idx + 1}`

        return {
          id: step.id || `parallel-${groupIdx}-${idx}`,
          installName,
          status: transformedStep.status,
          startedAt: step.created_at,
          completedAt: step.finished_at,
          executionTime: step.execution_time,
          steps: [transformedStep],
        }
      })
    } else {
      steps = sortedSteps.map(transformWorkflowStep)
    }

    const allSteps = parallelInstalls 
      ? parallelInstalls.flatMap(p => p.steps)
      : steps
    const status = determineStageStatus(allSteps)

    const executionTime = sortedSteps.reduce((sum, step) => {
      return sum + (step.execution_time || 0)
    }, 0)

    const startedAt = sortedSteps
      .filter(s => s.created_at)
      .sort((a, b) => new Date(a.created_at!).getTime() - new Date(b.created_at!).getTime())[0]?.created_at

    const completedAt = sortedSteps
      .filter(s => s.finished_at)
      .sort((a, b) => new Date(b.finished_at!).getTime() - new Date(a.finished_at!).getTime())[0]?.finished_at

    return {
      id: stageId,
      name: stageName,
      description: `${sortedSteps.length} ${isParallelStage ? 'parallel' : 'sequential'} operation${sortedSteps.length > 1 ? 's' : ''}`,
      status,
      startedAt,
      completedAt,
      executionTime: executionTime > 0 ? executionTime : undefined,
      steps,
      parallelInstalls,
      metadata: {
        groupIdx,
        stepCount: sortedSteps.length,
        isParallel: isParallelStage,
      },
    }
  })

  return stages
}
