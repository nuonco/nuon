'use client'

import { useState, useRef, useEffect } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { Button } from '@/components/common/Button'
import { Panel } from '@/components/surfaces/Panel'
import { cn } from '@/utils/classnames'

// Mock workflow stage and step types
type TWorkflowStageStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

interface IWorkflowStep {
  id: string
  name: string
  status: TWorkflowStageStatus
  executionTime?: number
  message?: string
  substeps?: IWorkflowStep[]
  logs?: string[]
  error?: string
}

interface IParallelInstallUpdate {
  id: string
  installName: string
  status: TWorkflowStageStatus
  startedAt?: string
  completedAt?: string
  executionTime?: number
  steps: IWorkflowStep[]
}

interface IWorkflowStage {
  id: string
  name: string
  description: string
  status: TWorkflowStageStatus
  startedAt?: string
  completedAt?: string
  executionTime?: number
  steps: IWorkflowStep[]
  parallelInstalls?: IParallelInstallUpdate[]
  metadata?: {
    commitHash?: string
    componentsChanged?: number
    installsAffected?: number
  }
}

interface IBranchWorkflowCanvas {
  branchId: string
}

// Mock data generator with sub-steps
const getMockWorkflowStages = (): IWorkflowStage[] => {
  return [
    {
      id: 'stage-1',
      name: 'Fetch Repository',
      description: 'Clone repository and checkout branch',
      status: 'completed',
      startedAt: new Date(Date.now() - 600000).toISOString(),
      completedAt: new Date(Date.now() - 580000).toISOString(),
      executionTime: 20000000000,
      metadata: {
        commitHash: 'abc1234',
      },
      steps: [
        {
          id: 'step-1-1',
          name: 'Clone repository',
          status: 'completed',
          executionTime: 8000000000,
        },
        {
          id: 'step-1-2',
          name: 'Checkout branch',
          status: 'completed',
          executionTime: 3000000000,
        },
        {
          id: 'step-1-3',
          name: 'Fetch dependencies',
          status: 'completed',
          executionTime: 9000000000,
        },
      ],
    },
    {
      id: 'stage-2',
      name: 'Build Config',
      description: 'Parse and validate application configuration',
      status: 'completed',
      startedAt: new Date(Date.now() - 580000).toISOString(),
      completedAt: new Date(Date.now() - 540000).toISOString(),
      executionTime: 40000000000,
      steps: [
        {
          id: 'step-2-1',
          name: 'Parse nuon.yaml',
          status: 'completed',
          executionTime: 1000000000,
          substeps: [
            {
              id: 'step-2-1-1',
              name: 'Load YAML file',
              status: 'completed',
              executionTime: 300000000,
            },
            {
              id: 'step-2-1-2',
              name: 'Validate YAML syntax',
              status: 'completed',
              executionTime: 200000000,
            },
            {
              id: 'step-2-1-3',
              name: 'Parse configuration structure',
              status: 'completed',
              executionTime: 500000000,
            },
          ],
          logs: [
            '2024-01-08 10:23:15 INFO Loading configuration from nuon.yaml',
            '2024-01-08 10:23:15 INFO YAML syntax validation passed',
            '2024-01-08 10:23:16 INFO Found 3 components in configuration',
            '2024-01-08 10:23:16 INFO Configuration parsed successfully',
          ],
        },
        {
          id: 'step-2-2',
          name: 'Validate configuration',
          status: 'completed',
          executionTime: 500000000,
        },
        {
          id: 'step-2-3',
          name: 'Resolve component dependencies',
          status: 'completed',
          executionTime: 2000000000,
          substeps: [
            {
              id: 'step-2-3-1',
              name: 'Build dependency graph',
              status: 'completed',
              executionTime: 800000000,
            },
            {
              id: 'step-2-3-2',
              name: 'Check for circular dependencies',
              status: 'completed',
              executionTime: 400000000,
            },
            {
              id: 'step-2-3-3',
              name: 'Resolve external dependencies',
              status: 'completed',
              executionTime: 800000000,
            },
          ],
        },
        {
          id: 'step-2-4',
          name: 'Create app config',
          status: 'completed',
          executionTime: 1500000000,
        },
      ],
    },
    {
      id: 'stage-3',
      name: 'Build Changed Components',
      description: 'Build and push container images for modified components',
      status: 'running',
      startedAt: new Date(Date.now() - 540000).toISOString(),
      metadata: {
        componentsChanged: 3,
      },
      steps: [
        {
          id: 'step-3-1',
          name: 'Detect changed components',
          status: 'completed',
          executionTime: 2000000000,
        },
        {
          id: 'step-3-2',
          name: 'Build component: api-service',
          status: 'completed',
          executionTime: 45000000000,
        },
        {
          id: 'step-3-3',
          name: 'Build component: web-frontend',
          status: 'running',
          message: 'Building Docker image...',
          substeps: [
            {
              id: 'step-3-3-1',
              name: 'Create build context',
              status: 'completed',
              executionTime: 1000000000,
            },
            {
              id: 'step-3-3-2',
              name: 'Execute Dockerfile',
              status: 'running',
            },
            {
              id: 'step-3-3-3',
              name: 'Tag image',
              status: 'pending',
            },
          ],
          logs: [
            '2024-01-08 10:25:30 INFO Starting build for web-frontend',
            '2024-01-08 10:25:31 INFO Base image pulled: node:20-alpine',
            '2024-01-08 10:25:32 INFO Installing dependencies...',
            '2024-01-08 10:25:45 INFO Building production bundle...',
          ],
        },
        {
          id: 'step-3-4',
          name: 'Build component: worker-service',
          status: 'pending',
        },
        {
          id: 'step-3-5',
          name: 'Push images to registry',
          status: 'pending',
        },
      ],
    },
    {
      id: 'stage-4',
      name: 'Update Installs',
      description: 'Deploy updated components to affected installs in parallel',
      status: 'pending',
      metadata: {
        installsAffected: 4,
      },
      steps: [
        {
          id: 'step-4-1',
          name: 'Identify affected installs',
          status: 'pending',
        },
      ],
      parallelInstalls: [
        {
          id: 'install-1',
          installName: 'Production Install',
          status: 'pending',
          steps: [
            {
              id: 'install-1-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-1-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-1-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-2',
          installName: 'Staging Install',
          status: 'pending',
          steps: [
            {
              id: 'install-2-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-2-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-2-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-3',
          installName: 'Dev Install',
          status: 'pending',
          steps: [
            {
              id: 'install-3-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-3-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-3-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-4',
          installName: 'QA Install',
          status: 'pending',
          steps: [
            {
              id: 'install-4-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-4-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-4-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
      ],
    },
    {
      id: 'stage-5',
      name: 'Update Installs',
      description: 'Deploy updated components to additional installs in parallel',
      status: 'pending',
      metadata: {
        installsAffected: 6,
      },
      steps: [
        {
          id: 'step-5-1',
          name: 'Identify additional installs',
          status: 'pending',
        },
      ],
      parallelInstalls: [
        {
          id: 'install-5',
          installName: 'Customer A Install',
          status: 'completed',
          steps: [
            {
              id: 'install-5-step-1',
              name: 'Generate deployment plan',
              status: 'completed',
            },
            {
              id: 'install-5-step-2',
              name: 'Apply plan to cluster',
              status: 'completed',
            },
            {
              id: 'install-5-step-3',
              name: 'Verify deployment health',
              status: 'completed',
            },
          ],
        },
        {
          id: 'install-6',
          installName: 'Customer B Install',
          status: 'failed',
          steps: [
            {
              id: 'install-6-step-1',
              name: 'Generate deployment plan',
              status: 'completed',
            },
            {
              id: 'install-6-step-2',
              name: 'Apply plan to cluster',
              status: 'failed',
              error: 'Connection timeout to cluster',
            },
            {
              id: 'install-6-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-7',
          installName: 'Customer C Install',
          status: 'completed',
          steps: [
            {
              id: 'install-7-step-1',
              name: 'Generate deployment plan',
              status: 'completed',
            },
            {
              id: 'install-7-step-2',
              name: 'Apply plan to cluster',
              status: 'completed',
            },
            {
              id: 'install-7-step-3',
              name: 'Verify deployment health',
              status: 'completed',
            },
          ],
        },
        {
          id: 'install-8',
          installName: 'Customer D Install',
          status: 'running',
          steps: [
            {
              id: 'install-8-step-1',
              name: 'Generate deployment plan',
              status: 'completed',
            },
            {
              id: 'install-8-step-2',
              name: 'Apply plan to cluster',
              status: 'running',
            },
            {
              id: 'install-8-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-9',
          installName: 'Customer E Install',
          status: 'pending',
          steps: [
            {
              id: 'install-9-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-9-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-9-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
        {
          id: 'install-10',
          installName: 'Customer F Install',
          status: 'pending',
          steps: [
            {
              id: 'install-10-step-1',
              name: 'Generate deployment plan',
              status: 'pending',
            },
            {
              id: 'install-10-step-2',
              name: 'Apply plan to cluster',
              status: 'pending',
            },
            {
              id: 'install-10-step-3',
              name: 'Verify deployment health',
              status: 'pending',
            },
          ],
        },
      ],
    },
  ]
}

// Side panel for step details
const StepDetailSidePanel = ({
  step,
  isOpen,
  onClose,
}: {
  step: IWorkflowStep | null
  isOpen: boolean
  onClose: () => void
}) => {
  if (!step) return null

  return (
    <Panel
      isVisible={isOpen}
      onClose={onClose}
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="List" size={20} />
          <div>
            <Text variant="base" weight="strong">
              Step Details
            </Text>
            <Text variant="subtext" theme="neutral">
              {step.name}
            </Text>
          </div>
        </div>
      }
      size="half"
    >
      <div className="flex flex-col gap-6">
        {/* Status card */}
        <Card>
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Icon
                variant={
                  step.status === 'completed'
                    ? 'CheckCircle'
                    : step.status === 'running'
                    ? 'CircleNotch'
                    : step.status === 'failed'
                    ? 'XCircle'
                    : 'Circle'
                }
                size={24}
                className={cn({
                  'text-green-600 dark:text-green-400': step.status === 'completed',
                  'text-blue-600 dark:text-blue-400 animate-spin': step.status === 'running',
                  'text-red-600 dark:text-red-400': step.status === 'failed',
                  'text-cool-grey-400 dark:text-dark-grey-500': step.status === 'pending',
                })}
              />
              <div>
                <Text variant="base" weight="strong">
                  {step.name}
                </Text>
                {step.message && (
                  <Text variant="subtext" theme="neutral" className="mt-1">
                    {step.message}
                  </Text>
                )}
              </div>
            </div>
            <Status status={step.status} variant="badge" />
          </div>

          {step.executionTime && (
            <div className="mt-4 pt-4 border-t">
              <div className="flex items-center gap-2">
                <Icon variant="Timer" size={16} className="text-cool-grey-500" />
                <Text variant="subtext" theme="neutral">
                  Execution Time:
                </Text>
                <Text variant="base" weight="strong">
                  <Duration variant="base" nanoseconds={step.executionTime} />
                </Text>
              </div>
            </div>
          )}
        </Card>

        {/* Error message */}
        {step.error && (
          <Card>
            <div className="flex items-start gap-3">
              <Icon
                variant="Warning"
                size={20}
                className="text-red-600 dark:text-red-400 mt-0.5"
              />
              <div className="flex-1">
                <Text variant="base" weight="strong" className="text-red-900 dark:text-red-200">
                  Error
                </Text>
                <Text variant="base" className="text-red-800 dark:text-red-300 mt-2">
                  {step.error}
                </Text>
              </div>
            </div>
          </Card>
        )}

        {/* Substeps */}
        {step.substeps && step.substeps.length > 0 && (
          <Card>
            <Text variant="base" weight="strong" className="mb-4">
              Substeps ({step.substeps.length})
            </Text>
            <div className="space-y-3">
              {step.substeps.map((substep, index) => (
                <div key={substep.id} className="flex gap-3">
                  <div className="flex flex-col items-center">
                    <Badge variant="default" size="sm" theme="neutral">
                      {index + 1}
                    </Badge>
                    {index < step.substeps!.length - 1 && (
                      <div className="w-px flex-1 bg-cool-grey-200 dark:bg-dark-grey-600 mt-2" />
                    )}
                  </div>
                  <div
                    className={cn(
                      'flex-1 p-3 rounded-md border transition-all',
                      {
                        'bg-green-50/50 border-green-200 dark:bg-green-950/20 dark:border-green-900':
                          substep.status === 'completed',
                        'bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900':
                          substep.status === 'running',
                        'bg-red-50/50 border-red-200 dark:bg-red-950/20 dark:border-red-900':
                          substep.status === 'failed',
                        'bg-cool-grey-50 border-cool-grey-200 dark:bg-dark-grey-800 dark:border-dark-grey-700':
                          substep.status === 'pending',
                      }
                    )}
                  >
                    <div className="flex items-center gap-3">
                      <Icon
                        variant={
                          substep.status === 'completed'
                            ? 'CheckCircle'
                            : substep.status === 'running'
                            ? 'CircleNotch'
                            : substep.status === 'failed'
                            ? 'XCircle'
                            : 'Circle'
                        }
                        size={16}
                        className={cn({
                          'text-green-600 dark:text-green-400': substep.status === 'completed',
                          'text-blue-600 dark:text-blue-400 animate-spin': substep.status === 'running',
                          'text-red-600 dark:text-red-400': substep.status === 'failed',
                          'text-cool-grey-400 dark:text-dark-grey-500': substep.status === 'pending',
                        })}
                      />
                      <Text variant="base" weight="normal" className="flex-1">
                        {substep.name}
                      </Text>
                      {substep.executionTime && (
                        <Text variant="subtext" theme="neutral">
                          <Duration variant="subtext" nanoseconds={substep.executionTime} />
                        </Text>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        )}

        {/* Logs */}
        {step.logs && step.logs.length > 0 && (
          <Card>
            <div className="flex items-center gap-2 mb-4">
              <Icon variant="Terminal" size={18} />
              <Text variant="base" weight="strong">
                Logs ({step.logs.length} lines)
              </Text>
            </div>
            <div className="p-4 bg-cool-grey-900 dark:bg-black rounded border border-cool-grey-700 overflow-x-auto">
              <div className="font-mono text-xs text-green-400 space-y-1">
                {step.logs.map((log, index) => (
                  <div key={index} className="whitespace-pre">{log}</div>
                ))}
              </div>
            </div>
          </Card>
        )}

        {/* Step ID for reference */}
        <Card>
          <Text variant="base" weight="strong" className="mb-2">
            Step ID
          </Text>
          <Badge variant="code" theme="neutral">
            {step.id}
          </Badge>
        </Card>
      </div>
    </Panel>
  )
}

// Simplified workflow stage card - CI/CD pipeline style
const WorkflowStageCard = ({
  stage,
  isSelected,
  onClick,
  isExpanded = false,
  onToggleExpand,
}: {
  stage: IWorkflowStage
  isSelected: boolean
  onClick: () => void
  isExpanded?: boolean
  onToggleExpand?: () => void
}) => {
  const getStatusIcon = () => {
    switch (stage.status) {
      case 'completed':
        return { icon: 'check', color: 'text-green-600 dark:text-green-400' }
      case 'failed':
        return { icon: 'times', color: 'text-red-600 dark:text-red-400' }
      case 'running':
        return {
          icon: 'CircleNotch',
          color: 'text-blue-600 dark:text-blue-400 animate-spin',
        }
      case 'pending':
      default:
        return { icon: 'Circle', color: 'text-cool-grey-400 dark:text-dark-grey-400' }
    }
  }

  const statusIconData = getStatusIcon()
  const hasParallelInstalls = stage.parallelInstalls && stage.parallelInstalls.length > 0
  const COLLAPSE_THRESHOLD = 4
  const shouldShowExpandButton = hasParallelInstalls && stage.parallelInstalls!.length > COLLAPSE_THRESHOLD

  // Special rendering for "Update Installs" stage with parallel installs
  if (hasParallelInstalls && stage.name === 'Update Installs') {
    const visibleInstalls = isExpanded 
      ? stage.parallelInstalls! 
      : stage.parallelInstalls!.slice(0, COLLAPSE_THRESHOLD)
    const hiddenCount = stage.parallelInstalls!.length - COLLAPSE_THRESHOLD

    return (
      <div className="flex flex-col gap-1.5">
        {visibleInstalls.map((install, idx) => {
          const status = install.status
          
          return (
            <button
              key={install.id}
              onClick={onClick}
              className={cn(
                'relative flex items-center gap-3 px-4 py-2',
                'min-w-[280px] rounded border-2 transition-all duration-200',
                'cursor-pointer select-none',
                'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                'hover:shadow-md',
                {
                  // Completed state
                  'border-green-400 bg-green-50 dark:bg-green-950/30':
                    status === 'completed' && !isSelected,
                  'border-green-500 bg-green-100 dark:bg-green-900/40 shadow-lg ring-2 ring-green-500':
                    status === 'completed' && isSelected,
                  // Running state
                  'border-blue-400 bg-blue-50 dark:bg-blue-950/30':
                    status === 'running' && !isSelected,
                  'border-blue-500 bg-blue-100 dark:bg-blue-900/40 shadow-lg ring-2 ring-blue-500':
                    status === 'running' && isSelected,
                  // Failed state
                  'border-red-400 bg-red-50 dark:bg-red-950/30':
                    status === 'failed' && !isSelected,
                  'border-red-500 bg-red-100 dark:bg-red-900/40 shadow-lg ring-2 ring-red-500':
                    status === 'failed' && isSelected,
                  // Pending state
                  'border-cool-grey-300 bg-cool-grey-50 dark:bg-dark-grey-800/50':
                    status === 'pending' && !isSelected,
                  'border-cool-grey-400 bg-cool-grey-100 dark:bg-dark-grey-700/60 shadow-lg ring-2 ring-cool-grey-400':
                    status === 'pending' && isSelected,
                }
              )}
            >
              {/* Status icon - compact */}
              <Icon
                variant={
                  status === 'completed'
                    ? 'Check'
                    : status === 'running'
                    ? 'CircleNotch'
                    : status === 'failed'
                    ? 'X'
                    : 'Circle'
                }
                size={16}
                className={cn({
                  'text-green-600 dark:text-green-400': status === 'completed',
                  'text-blue-600 dark:text-blue-400 animate-spin': status === 'running',
                  'text-red-600 dark:text-red-400': status === 'failed',
                  'text-cool-grey-400 dark:text-dark-grey-400': status === 'pending',
                })}
              />

              {/* Install info */}
              <div className="flex flex-col items-start flex-1 min-w-0">
                <Text variant="subtext" weight="normal" className="truncate">
                  Update Installs
                </Text>
                <Text variant="caption" theme="neutral" className="font-mono text-xs truncate max-w-full">
                  {install.id}
                </Text>
              </div>

              {/* Selected indicator for first row only */}
              {isSelected && idx === 0 && (
                <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
                  <Icon
                    variant="caret-down"
                    size={20}
                    className="text-blue-600 dark:text-blue-400"
                  />
                </div>
              )}
            </button>
          )
        })}

        {/* Expand/Collapse button for stages with >4 installs */}
        {shouldShowExpandButton && (
          <button
            onClick={(e) => {
              e.stopPropagation()
              onToggleExpand?.()
            }}
            className={cn(
              'flex items-center justify-center gap-2 px-4 py-2',
              'min-w-[280px] rounded border-2 transition-all duration-200',
              'cursor-pointer select-none',
              'border-cool-grey-300 bg-cool-grey-100 hover:bg-cool-grey-200',
              'dark:border-dark-grey-600 dark:bg-dark-grey-700 dark:hover:bg-dark-grey-600',
              'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
            )}
          >
            <Icon
              variant={isExpanded ? 'chevron-up' : 'chevron-down'}
              size={16}
              className="text-cool-grey-600 dark:text-cool-grey-400"
            />
            <Text variant="subtext" weight="normal" className="text-cool-grey-700 dark:text-cool-grey-300">
              {isExpanded ? 'Collapse' : `+${hiddenCount} more`}
            </Text>
          </button>
        )}
      </div>
    )
  }

  // Standard single-card rendering for other stages
  return (
    <button
      onClick={onClick}
      className={cn(
        'relative flex flex-col items-center justify-center',
        'min-w-[200px] h-[140px] p-4 rounded-lg border-2 transition-all duration-200',
        'cursor-pointer select-none',
        'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
        'hover:shadow-md',
        {
          // Completed state
          'border-green-500 bg-green-50 dark:bg-green-950/30':
            stage.status === 'completed' && !isSelected,
          'border-green-600 bg-green-100 dark:bg-green-900/40 shadow-lg ring-2 ring-green-500':
            stage.status === 'completed' && isSelected,
          // Running state
          'border-blue-500 bg-blue-50 dark:bg-blue-950/30':
            stage.status === 'running' && !isSelected,
          'border-blue-600 bg-blue-100 dark:bg-blue-900/40 shadow-lg ring-2 ring-blue-500':
            stage.status === 'running' && isSelected,
          // Failed state
          'border-red-500 bg-red-50 dark:bg-red-950/30':
            stage.status === 'failed' && !isSelected,
          'border-red-600 bg-red-100 dark:bg-red-900/40 shadow-lg ring-2 ring-red-500':
            stage.status === 'failed' && isSelected,
          // Pending state
          'border-cool-grey-300 bg-cool-grey-50 dark:bg-dark-grey-800/50':
            stage.status === 'pending' && !isSelected,
          'border-cool-grey-400 bg-cool-grey-100 dark:bg-dark-grey-700/60 shadow-lg ring-2 ring-cool-grey-400':
            stage.status === 'pending' && isSelected,
        }
      )}
    >
      {/* Status icon - large and prominent */}
      <div className="mb-3">
        <Icon
          variant={statusIconData.icon as any}
          size={32}
          className={statusIconData.color}
        />
      </div>

      {/* Stage name */}
      <Text
        variant="base"
        weight="strong"
        className="text-center leading-tight"
      >
        {stage.name}
      </Text>

      {/* Selected indicator */}
      {isSelected && (
        <div className="absolute -bottom-2 left-1/2 -translate-x-1/2">
          <Icon
            variant="caret-down"
            size={20}
            className="text-blue-600 dark:text-blue-400"
          />
        </div>
      )}
    </button>
  )
}

// Connector arrow component - CI pipeline style
const StageConnector = ({ isActive }: { isActive: boolean }) => {
  return (
    <div className="flex items-center justify-center px-6">
      <div className="flex items-center gap-1">
        <div
          className={cn('h-0.5 w-8 transition-colors duration-200', {
            'bg-green-500 dark:bg-green-400': isActive,
            'bg-cool-grey-300 dark:bg-dark-grey-600': !isActive,
          })}
        />
        <Icon
          variant="caret-right"
          size={16}
          className={cn('transition-colors duration-200', {
            'text-green-500 dark:text-green-400': isActive,
            'text-cool-grey-400 dark:text-dark-grey-500': !isActive,
          })}
        />
      </div>
    </div>
  )
}

// Collapsible step detail row component
const CollapsibleStepDetailRow = ({ 
  step,
  isExpanded,
  onToggle,
  onOpenPanel,
}: { 
  step: IWorkflowStep
  isExpanded: boolean
  onToggle: () => void
  onOpenPanel?: (step: IWorkflowStep) => void
}) => {
  const getStatusIcon = (status: TWorkflowStageStatus) => {
    switch (status) {
      case 'completed':
        return 'CheckCircle'
      case 'running':
        return 'CircleNotch'
      case 'failed':
        return 'XCircle'
      case 'cancelled':
        return 'ban'
      case 'pending':
      default:
        return 'Circle'
    }
  }

  const hasDetails = step.substeps || step.logs || step.error

  return (
    <div
      className={cn(
        'flex flex-col rounded-md border transition-all duration-200',
        {
          'bg-green-50/50 border-green-200 dark:bg-green-950/20 dark:border-green-900':
            step.status === 'completed',
          'bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900':
            step.status === 'running',
          'bg-red-50/50 border-red-200 dark:bg-red-950/20 dark:border-red-900':
            step.status === 'failed',
          'bg-cool-grey-50 border-cool-grey-200 dark:bg-dark-grey-800 dark:border-dark-grey-700':
            step.status === 'pending',
        }
      )}
    >
      {/* Collapsed view - always visible */}
      <button
        onClick={onToggle}
        className={cn(
          'flex items-start gap-4 p-3 text-left w-full transition-colors',
          hasDetails && 'hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer',
          !hasDetails && 'cursor-default'
        )}
        disabled={!hasDetails}
      >
        <div className="mt-0.5">
          <Icon
            variant={getStatusIcon(step.status)}
            size={18}
            className={cn({
              'text-green-600 dark:text-green-400': step.status === 'completed',
              'text-blue-600 dark:text-blue-400 animate-spin':
                step.status === 'running',
              'text-red-600 dark:text-red-400': step.status === 'failed',
              'text-cool-grey-400 dark:text-dark-grey-500':
                step.status === 'pending',
            })}
          />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Text variant="base" weight="normal">
              {step.name}
            </Text>
            {hasDetails && (
              <Icon
                variant="chevron-down"
                size={14}
                className={cn(
                  'text-cool-grey-500 transition-transform duration-200',
                  isExpanded && 'rotate-180'
                )}
              />
            )}
          </div>
          {step.message && (
            <Text variant="subtext" theme="neutral" className="mt-1">
              {step.message}
            </Text>
          )}
        </div>
        <div className="flex items-center gap-2">
          {step.executionTime && step.status === 'completed' && (
            <Badge variant="default" size="sm" theme="success">
              <Duration variant="subtext" nanoseconds={step.executionTime} />
            </Badge>
          )}
          <Status status={step.status} variant="badge" />
          {onOpenPanel && (hasDetails || step.logs || step.error) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onOpenPanel(step)
              }}
              className="!p-1"
            >
              <Icon variant="arrow-right" size={14} />
            </Button>
          )}
        </div>
      </button>

      {/* Expanded view - details */}
      {hasDetails && isExpanded && (
        <div className="px-3 pb-3 pt-0 border-t border-current/10">
          <div className="pl-7 space-y-3 mt-3">
            {/* Error message */}
            {step.error && (
              <div className="p-3 bg-red-100 dark:bg-red-950/40 rounded-md border border-red-300 dark:border-red-900">
                <div className="flex items-start gap-2">
                  <Icon
                    variant="Warning"
                    size={16}
                    className="text-red-600 dark:text-red-400 mt-0.5"
                  />
                  <div>
                    <Text variant="base" weight="strong" className="text-red-900 dark:text-red-200">
                      Error
                    </Text>
                    <Text variant="subtext" className="text-red-800 dark:text-red-300 mt-1">
                      {step.error}
                    </Text>
                  </div>
                </div>
              </div>
            )}

            {/* Substeps */}
            {step.substeps && step.substeps.length > 0 && (
              <div>
                <Text variant="base" weight="normal" className="mb-2">
                  Substeps:
                </Text>
                <div className="space-y-2">
                  {step.substeps.map((substep) => (
                    <div
                      key={substep.id}
                      className="flex items-center gap-3 p-2 bg-white/50 dark:bg-black/20 rounded border border-current/10"
                    >
                      <Icon
                        variant={getStatusIcon(substep.status)}
                        size={14}
                        className={cn({
                          'text-green-600 dark:text-green-400': substep.status === 'completed',
                          'text-blue-600 dark:text-blue-400 animate-spin': substep.status === 'running',
                          'text-red-600 dark:text-red-400': substep.status === 'failed',
                          'text-cool-grey-400 dark:text-dark-grey-500': substep.status === 'pending',
                        })}
                      />
                      <Text variant="subtext" className="flex-1">
                        {substep.name}
                      </Text>
                      {substep.executionTime && (
                        <Text variant="subtext" theme="neutral">
                          <Duration variant="subtext" nanoseconds={substep.executionTime} />
                        </Text>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Logs */}
            {step.logs && step.logs.length > 0 && (
              <div>
                <Text variant="base" weight="normal" className="mb-2">
                  Logs:
                </Text>
                <div className="p-3 bg-cool-grey-900 dark:bg-black rounded border border-cool-grey-700">
                  <div className="font-mono text-xs text-green-400 space-y-1">
                    {step.logs.map((log, index) => (
                      <div key={index}>{log}</div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// Detail section component with enhanced summary
const StageDetailSection = ({ 
  stage,
  expandedSteps,
  onToggleStep,
  onOpenStepPanel,
}: { 
  stage: IWorkflowStage
  expandedSteps: Set<string>
  onToggleStep: (stepId: string) => void
  onOpenStepPanel: (step: IWorkflowStep) => void
}) => {
  // Calculate summary metrics including parallel installs
  let allSteps = [...stage.steps]
  if (stage.parallelInstalls) {
    stage.parallelInstalls.forEach(install => {
      allSteps = allSteps.concat(install.steps)
    })
  }
  
  const completedSteps = allSteps.filter((s) => s.status === 'completed').length
  const runningSteps = allSteps.filter((s) => s.status === 'running').length
  const failedSteps = allSteps.filter((s) => s.status === 'failed').length
  const pendingSteps = allSteps.filter((s) => s.status === 'pending').length

  return (
    <div className="flex flex-col gap-6">
      {/* Summary Card */}
      <Card>
        <div className="flex flex-col gap-4">
          {/* Header with status */}
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <Text variant="h4" weight="strong">
                  {stage.name}
                </Text>
                <Status status={stage.status} variant="badge" />
              </div>
              <Text variant="base" theme="neutral">
                {stage.description}
              </Text>
            </div>
          </div>

          {/* Key metrics grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t">
            {/* Total execution time */}
            {stage.executionTime && (
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Total Time
                </Text>
                <div className="flex items-center gap-2">
                  <Icon variant="timer" size={16} />
                  <Text variant="base" weight="strong">
                    <Duration variant="base" nanoseconds={stage.executionTime} />
                  </Text>
                </div>
              </div>
            )}

            {/* Components changed */}
            {stage.metadata?.componentsChanged !== undefined && (
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Components Built
                </Text>
                <div className="flex items-center gap-2">
                  <Icon variant="package" size={16} />
                  <Text variant="base" weight="strong">
                    {stage.metadata.componentsChanged}
                  </Text>
                </div>
              </div>
            )}

            {/* Installs affected */}
            {stage.metadata?.installsAffected !== undefined && (
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Installs Updated
                </Text>
                <div className="flex items-center gap-2">
                  <Icon variant="cloud" size={16} />
                  <Text variant="base" weight="strong">
                    {stage.metadata.installsAffected}
                  </Text>
                </div>
              </div>
            )}

            {/* Step progress */}
            <div className="flex flex-col gap-1">
              <Text variant="subtext" theme="neutral">
                Step Progress
              </Text>
              <div className="flex items-center gap-2">
                <Icon variant="list" size={16} />
                <Text variant="base" weight="strong">
                  {completedSteps} / {stage.steps.length}
                </Text>
              </div>
            </div>
          </div>

          {/* Timestamps */}
          <div className="flex flex-wrap gap-6 pt-4 border-t">
            {stage.startedAt && (
              <div className="flex items-center gap-2">
                <Icon variant="play" size={14} className="text-cool-grey-500" />
                <Text variant="subtext" theme="neutral">
                  Started:{' '}
                  <span className="font-medium">
                    {new Date(stage.startedAt).toLocaleString()}
                  </span>
                </Text>
              </div>
            )}
            {stage.completedAt && (
              <div className="flex items-center gap-2">
                <Icon
                  variant="check-circle"
                  size={14}
                  className="text-green-600"
                />
                <Text variant="subtext" theme="neutral">
                  Completed:{' '}
                  <span className="font-medium">
                    {new Date(stage.completedAt).toLocaleString()}
                  </span>
                </Text>
              </div>
            )}
            {stage.metadata?.commitHash && (
              <div className="flex items-center gap-2">
                <Icon variant="git-commit" size={14} />
                <Badge variant="code" size="sm" theme="default">
                  {stage.metadata.commitHash}
                </Badge>
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Steps Detail Card */}
      <Card>
        <div className="flex flex-col gap-6">
          {/* Steps header with counts */}
          <div className="flex items-center justify-between">
            <Text variant="base" weight="strong">
              Detailed Steps ({stage.steps.length})
            </Text>
            <div className="flex items-center gap-4">
              {completedSteps > 0 && (
                <div className="flex items-center gap-2">
                  <Icon
                    variant="check-circle"
                    size={14}
                    className="text-green-600"
                  />
                  <Text variant="subtext" theme="neutral">
                    {completedSteps} completed
                  </Text>
                </div>
              )}
              {runningSteps > 0 && (
                <div className="flex items-center gap-2">
                  <Icon
                    variant="circle-notch"
                    size={14}
                    className="text-blue-600 animate-spin"
                  />
                  <Text variant="subtext" theme="neutral">
                    {runningSteps} running
                  </Text>
                </div>
              )}
              {failedSteps > 0 && (
                <div className="flex items-center gap-2">
                  <Icon
                    variant="times-circle"
                    size={14}
                    className="text-red-600"
                  />
                  <Text variant="subtext" theme="neutral">
                    {failedSteps} failed
                  </Text>
                </div>
              )}
              {pendingSteps > 0 && (
                <div className="flex items-center gap-2">
                  <Icon
                    variant="circle"
                    size={14}
                    className="text-cool-grey-400"
                  />
                  <Text variant="subtext" theme="neutral">
                    {pendingSteps} pending
                  </Text>
                </div>
              )}
            </div>
          </div>

          {/* Steps list */}
          <div className="flex flex-col gap-3">
            {stage.steps.map((step, index) => (
              <div key={step.id} className="flex gap-3">
                <div className="flex flex-col items-center">
                  <Badge variant="default" size="sm" theme="neutral">
                    {index + 1}
                  </Badge>
                  {index < stage.steps.length - 1 && (
                    <div className="w-px flex-1 bg-cool-grey-200 dark:bg-dark-grey-600 mt-2" />
                  )}
                </div>
                <div className="flex-1">
                  <CollapsibleStepDetailRow 
                    step={step}
                    isExpanded={expandedSteps.has(step.id)}
                    onToggle={() => onToggleStep(step.id)}
                    onOpenPanel={onOpenStepPanel}
                  />
                </div>
              </div>
            ))}
          </div>

          {/* Parallel installs section */}
          {stage.parallelInstalls && stage.parallelInstalls.length > 0 && (
            <div className="mt-6 pt-6 border-t">
              <div className="flex items-center gap-3 mb-4">
                <Icon variant="layer-group" size={20} className="text-blue-600 dark:text-blue-400" />
                <Text variant="base" weight="strong">
                  Parallel Install Updates ({stage.parallelInstalls.length})
                </Text>
                <Badge variant="default" theme="info" size="sm">
                  Running in parallel
                </Badge>
              </div>
              
              {/* Grid layout for parallel installs */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {stage.parallelInstalls.map((install) => {
                  const installCompleted = install.steps.filter(s => s.status === 'completed').length
                  const installTotal = install.steps.length
                  
                  return (
                    <div
                      key={install.id}
                      className={cn(
                        'p-4 rounded-lg border-2 transition-all',
                        {
                          'border-green-300 bg-green-50 dark:bg-green-950/20': install.status === 'completed',
                          'border-blue-300 bg-blue-50 dark:bg-blue-950/20': install.status === 'running',
                          'border-red-300 bg-red-50 dark:bg-red-950/20': install.status === 'failed',
                          'border-cool-grey-300 bg-cool-grey-50 dark:bg-dark-grey-800': install.status === 'pending',
                        }
                      )}
                    >
                      {/* Install header */}
                      <div className="flex items-start justify-between gap-3 mb-4">
                        <div className="flex items-center gap-2 flex-1">
                          <Icon variant="cloud" size={16} />
                          <Text variant="base" weight="strong">
                            {install.installName}
                          </Text>
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge variant="default" size="sm" theme="neutral">
                            {installCompleted}/{installTotal}
                          </Badge>
                          <Status status={install.status} variant="badge" />
                        </div>
                      </div>

                      {/* Install timing */}
                      {(install.startedAt || install.executionTime) && (
                        <div className="flex gap-4 mb-4 text-xs">
                          {install.startedAt && (
                            <Text variant="subtext" theme="neutral">
                              Started: {new Date(install.startedAt).toLocaleTimeString()}
                            </Text>
                          )}
                          {install.executionTime && (
                            <Text variant="subtext" theme="neutral">
                              Duration: <Duration variant="subtext" nanoseconds={install.executionTime} />
                            </Text>
                          )}
                        </div>
                      )}

                      {/* Install steps */}
                      <div className="space-y-2">
                        {install.steps.map((step, stepIndex) => (
                          <div key={step.id} className="flex gap-2">
                            <div className="flex flex-col items-center pt-1">
                              <Badge variant="default" size="sm" theme="neutral" className="text-xs w-6 h-6 flex items-center justify-center">
                                {stepIndex + 1}
                              </Badge>
                              {stepIndex < install.steps.length - 1 && (
                                <div className="w-px flex-1 bg-cool-grey-200 dark:bg-dark-grey-600 mt-1" />
                              )}
                            </div>
                            <div className="flex-1 min-w-0">
                              <CollapsibleStepDetailRow
                                step={step}
                                isExpanded={expandedSteps.has(step.id)}
                                onToggle={() => onToggleStep(step.id)}
                                onOpenPanel={onOpenStepPanel}
                              />
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}

// Draggable canvas component
const DraggableCanvas = ({
  children,
}: {
  children: React.ReactNode
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [startX, setStartX] = useState(0)
  const [scrollLeft, setScrollLeft] = useState(0)

  const handleMouseDown = (e: React.MouseEvent) => {
    if (!containerRef.current) return
    setIsDragging(true)
    setStartX(e.pageX - containerRef.current.offsetLeft)
    setScrollLeft(containerRef.current.scrollLeft)
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging || !containerRef.current) return
    e.preventDefault()
    const x = e.pageX - containerRef.current.offsetLeft
    const walk = (x - startX) * 1.5 // Scroll speed multiplier
    containerRef.current.scrollLeft = scrollLeft - walk
  }

  const handleMouseLeave = () => {
    setIsDragging(false)
  }

  // Center the canvas on mount
  useEffect(() => {
    if (containerRef.current && contentRef.current) {
      const containerWidth = containerRef.current.offsetWidth
      const contentWidth = contentRef.current.offsetWidth
      const centerPosition = (contentWidth - containerWidth) / 2
      containerRef.current.scrollLeft = Math.max(0, centerPosition)
    }
  }, [])

  return (
    <div
      ref={containerRef}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      className={cn(
        'relative overflow-x-auto overflow-y-hidden',
        'bg-cool-grey-50 dark:bg-dark-grey-900 rounded-lg border-2 border-cool-grey-200 dark:border-dark-grey-700',
        'py-12 px-8',
        {
          'cursor-grabbing': isDragging,
          'cursor-grab': !isDragging,
        }
      )}
      style={{
        scrollbarWidth: 'none', // Firefox
        msOverflowStyle: 'none', // IE/Edge
      }}
    >
      {/* Hide scrollbar for Chrome/Safari */}
      <style jsx>{`
        div::-webkit-scrollbar {
          display: none;
        }
      `}</style>

      <div ref={contentRef} className="inline-flex items-center gap-0 min-h-[180px]">
        {children}
      </div>

      {/* Drag hint */}
      {!isDragging && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-1.5 bg-cool-grey-800/80 dark:bg-dark-grey-700/80 rounded-full backdrop-blur-sm">
          <Icon
            variant="arrows-alt-h"
            size={14}
            className="text-cool-grey-300 dark:text-dark-grey-300"
          />
          <Text variant="subtext" className="text-cool-grey-300 dark:text-dark-grey-300">
            Drag to navigate pipeline
          </Text>
        </div>
      )}
    </div>
  )
}

export const BranchWorkflowCanvas = ({ branchId }: IBranchWorkflowCanvas) => {
  const stages = getMockWorkflowStages()
  const [selectedStage, setSelectedStage] = useState<IWorkflowStage>(stages[0])
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set())
  const [selectedStep, setSelectedStep] = useState<IWorkflowStep | null>(null)
  const [isPanelOpen, setIsPanelOpen] = useState(false)
  const [expandedParallelStages, setExpandedParallelStages] = useState<Set<string>>(new Set())

  const handleToggleStep = (stepId: string) => {
    setExpandedSteps(prev => {
      const next = new Set(prev)
      if (next.has(stepId)) {
        next.delete(stepId)
      } else {
        next.add(stepId)
      }
      return next
    })
  }

  const handleOpenStepPanel = (step: IWorkflowStep) => {
    setSelectedStep(step)
    setIsPanelOpen(true)
  }

  const handleClosePanel = () => {
    setIsPanelOpen(false)
    setTimeout(() => setSelectedStep(null), 300)
  }

  const handleToggleParallelStage = (stageId: string) => {
    setExpandedParallelStages(prev => {
      const next = new Set(prev)
      if (next.has(stageId)) {
        next.delete(stageId)
      } else {
        next.add(stageId)
      }
      return next
    })
  }

  return (
    <div className="w-full mt-8 flex flex-col gap-8">
      {/* Side panel for step details */}
      <StepDetailSidePanel
        step={selectedStep}
        isOpen={isPanelOpen}
        onClose={handleClosePanel}
      />

      {/* Canvas section header */}
      <div>
        <Text variant="h4" weight="strong">
          Workflow Pipeline
        </Text>
        <Text variant="base" theme="neutral" className="mt-1">
          Drag the canvas to navigate. Click any stage to view details below. Click step arrows to view logs and details.
        </Text>
      </div>

      {/* Draggable canvas */}
      <DraggableCanvas>
        {stages.map((stage, index) => (
          <div key={stage.id} className="flex items-center">
            <WorkflowStageCard
              stage={stage}
              isSelected={selectedStage.id === stage.id}
              onClick={() => setSelectedStage(stage)}
              isExpanded={expandedParallelStages.has(stage.id)}
              onToggleExpand={() => handleToggleParallelStage(stage.id)}
            />
            {index < stages.length - 1 && (
              <StageConnector
                isActive={
                  stage.status === 'completed' ||
                  (stage.status === 'running' &&
                    stages[index + 1]?.status === 'pending')
                }
              />
            )}
          </div>
        ))}
      </DraggableCanvas>

      {/* Detail section */}
      {selectedStage && (
        <StageDetailSection 
          stage={selectedStage}
          expandedSteps={expandedSteps}
          onToggleStep={handleToggleStep}
          onOpenStepPanel={handleOpenStepPanel}
        />
      )}

      {/* Mock data notice */}
      <div className="p-4 bg-blue-50 dark:bg-blue-950/20 border border-blue-400 dark:border-blue-500/40 rounded-md">
        <div className="flex items-start gap-3">
          <Icon
            variant="info"
            size={20}
            className="text-blue-600 dark:text-blue-400 mt-0.5"
          />
          <div>
            <Text variant="base" weight="strong" theme="info">
              Mock Data Preview
            </Text>
            <Text variant="subtext" theme="info" className="mt-1">
              This is a preview with mock workflow data. In production, this
              will display real workflow stages and their status for branch{' '}
              <Badge variant="code" size="sm" theme="info">
                {branchId}
              </Badge>
            </Text>
          </div>
        </div>
      </div>
    </div>
  )
}