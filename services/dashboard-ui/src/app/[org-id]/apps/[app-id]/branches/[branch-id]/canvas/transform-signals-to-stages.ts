import type { TQueueSignal } from '@/lib/ctl-api/queues/get-queue-signals'

export type TWorkflowStageStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

export interface IWorkflowStep {
  id: string
  name: string
  status: TWorkflowStageStatus
  executionTime?: number
  message?: string
  substeps?: IWorkflowStep[]
  logs?: string[]
  error?: string
}

export interface IWorkflowStage {
  id: string
  name: string
  description: string
  status: TWorkflowStageStatus
  startedAt?: string
  completedAt?: string
  executionTime?: number
  steps: IWorkflowStep[]
  metadata?: {
    commitHash?: string
    componentsChanged?: number
    installsAffected?: number
  }
}

interface IStageGroup {
  stageId: string
  stageName: string
  description: string
  signals: TQueueSignal[]
  allCompleted: boolean
  metadata?: Record<string, any>
}

// Map signal types to logical stage names
const getStageKeyFromSignalType = (signalType: string): string => {
  // Group similar signal types into stages
  if (signalType.includes('build') || signalType.includes('compile')) {
    return 'build'
  }
  if (signalType.includes('deploy') || signalType.includes('install')) {
    return 'deploy'
  }
  if (signalType.includes('test') || signalType.includes('validate')) {
    return 'test'
  }
  if (signalType.includes('config') || signalType.includes('setup')) {
    return 'config'
  }
  // Default: use the signal type as-is
  return signalType.toLowerCase().replace(/[^a-z0-9]/g, '-')
}

const getStageName = (stageKey: string): string => {
  const names: Record<string, string> = {
    build: 'Build Components',
    deploy: 'Deploy to Installs',
    test: 'Run Tests',
    config: 'Configure Environment',
  }
  return names[stageKey] || stageKey
}

const getStageDescription = (stageKey: string): string => {
  const descriptions: Record<string, string> = {
    build: 'Build and package application components',
    deploy: 'Deploy application to target installs',
    test: 'Run validation and integration tests',
    config: 'Configure application settings and environment',
  }
  return descriptions[stageKey] || `Execute ${stageKey} operations`
}

const groupSignalsByStageType = (signals: TQueueSignal[]): IStageGroup[] => {
  const stageMap = new Map<string, TQueueSignal[]>()

  signals.forEach((signal) => {
    const stageKey = getStageKeyFromSignalType(signal.type)
    if (!stageMap.has(stageKey)) {
      stageMap.set(stageKey, [])
    }
    stageMap.get(stageKey)!.push(signal)
  })

  return Array.from(stageMap.entries()).map(([stageKey, groupSignals]) => ({
    stageId: stageKey,
    stageName: getStageName(stageKey),
    description: getStageDescription(stageKey),
    signals: groupSignals,
    allCompleted: groupSignals.every((s) => s.status.status === 'completed'),
    metadata: extractStageMetadata(groupSignals),
  }))
}

const extractStageMetadata = (signals: TQueueSignal[]): Record<string, any> => {
  const metadata: Record<string, any> = {}

  // Extract common metadata from signals
  signals.forEach((signal) => {
    if (signal.signal) {
      // Look for common fields in signal data
      if (signal.signal.commit_hash) {
        metadata.commitHash = signal.signal.commit_hash
      }
      if (signal.signal.components_changed) {
        metadata.componentsChanged = signal.signal.components_changed
      }
      if (signal.signal.installs_affected) {
        metadata.installsAffected = signal.signal.installs_affected
      }
    }
  })

  return metadata
}

const deriveStageStatus = (signals: TQueueSignal[]): TWorkflowStageStatus => {
  if (signals.some((s) => s.status.status === 'failed')) return 'failed'
  if (signals.some((s) => s.status.status === 'cancelled')) return 'cancelled'
  if (signals.some((s) => s.status.status === 'running')) return 'running'
  if (signals.every((s) => s.status.status === 'completed')) return 'completed'
  return 'pending'
}

const calculateExecutionTime = (signals: TQueueSignal[]): number | undefined => {
  if (signals.length === 0) return undefined

  const startTimes = signals
    .map((s) => new Date(s.created_at).getTime())
    .filter((t) => !isNaN(t))

  const endTimes = signals
    .map((s) => new Date(s.updated_at).getTime())
    .filter((t) => !isNaN(t))

  if (startTimes.length === 0 || endTimes.length === 0) return undefined

  const earliestStart = Math.min(...startTimes)
  const latestEnd = Math.max(...endTimes)

  // Return time in nanoseconds (for Duration component)
  return (latestEnd - earliestStart) * 1_000_000
}

const calculateSignalExecutionTime = (signal: TQueueSignal): number | undefined => {
  const start = new Date(signal.created_at).getTime()
  const end = new Date(signal.updated_at).getTime()

  if (isNaN(start) || isNaN(end)) return undefined

  // Return time in nanoseconds
  return (end - start) * 1_000_000
}

const transformSignalsToSteps = (signals: TQueueSignal[]): IWorkflowStep[] => {
  return signals.map((signal) => ({
    id: signal.id,
    name: signal.type,
    status: signal.status.status as TWorkflowStageStatus,
    executionTime: calculateSignalExecutionTime(signal),
    message: signal.status.message,
    error: signal.status.error || signal.status.user_error,
    substeps: [],
    logs: [],
  }))
}

export const transformSignalsToStages = (
  signals: TQueueSignal[]
): IWorkflowStage[] => {
  if (signals.length === 0) return []

  // Group signals by stage type
  const stageGroups = groupSignalsByStageType(signals)

  return stageGroups.map((group) => ({
    id: group.stageId,
    name: group.stageName,
    description: group.description,
    status: deriveStageStatus(group.signals),
    startedAt: group.signals[0]?.created_at,
    completedAt: group.allCompleted
      ? group.signals[group.signals.length - 1]?.updated_at
      : undefined,
    executionTime: calculateExecutionTime(group.signals),
    steps: transformSignalsToSteps(group.signals),
    metadata: group.metadata,
  }))
}
