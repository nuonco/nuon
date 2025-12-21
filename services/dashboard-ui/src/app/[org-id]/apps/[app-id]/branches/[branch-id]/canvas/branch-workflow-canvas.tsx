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
import type { TWorkflow } from '@/types'
import {
  transformWorkflowToStages,
  type IWorkflowStep,
  type IWorkflowStage,
  type TWorkflowStageStatus,
  type IParallelInstallUpdate,
} from '@/lib/workflows/transform-workflow-to-stages'

interface IBranchWorkflowCanvas {
  workflow: TWorkflow
  branchId?: string
  appId?: string
  orgId?: string
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
    // Special icon for approval stage
    if (stage.status === 'pending' && stage.name === 'Pending Approval') {
      return { icon: 'Clock', color: 'text-amber-600 dark:text-amber-400' }
    }
    
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

  // Special rendering for stages with parallel operations (builds or installs)
  const isParallelStage = hasParallelInstalls && (
    stage.name === 'Update Installs' || 
    stage.name === 'Build Components' || 
    stage.name === 'Deploy Batch 1' || 
    stage.name === 'Deploy Batch 2' ||
    stage.name === 'Deploy to Installs'
  )
  
  if (isParallelStage) {
    const visibleInstalls = isExpanded 
      ? stage.parallelInstalls! 
      : stage.parallelInstalls!.slice(0, COLLAPSE_THRESHOLD)
    const hiddenCount = stage.parallelInstalls!.length - COLLAPSE_THRESHOLD

    return (
      <div className="relative flex flex-col gap-0.5">
        {visibleInstalls.map((install, idx) => {
          const status = install.status
          // Create stacking effect with increasing offset and shadow depth
          const stackOffset = idx * 4 // 4px stagger per card
          
          return (
            <button
              key={install.id}
              onClick={onClick}
              style={{
                marginLeft: `${stackOffset}px`,
                zIndex: visibleInstalls.length - idx, // Front cards have higher z-index
              }}
              className={cn(
                'relative flex items-center gap-3 px-4 py-2.5',
                'min-w-[280px] rounded-md border-2 transition-all duration-300',
                'cursor-pointer select-none',
                'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                // Enhanced hover with lift effect
                'hover:-translate-y-1 hover:shadow-xl hover:scale-[1.02]',
                {
                  // Completed state with layered shadows
                  'border-green-400 bg-green-50 dark:bg-green-950/30 shadow-md':
                    status === 'completed' && !isSelected,
                  'border-green-500 bg-green-100 dark:bg-green-900/40 shadow-2xl ring-2 ring-green-500':
                    status === 'completed' && isSelected,
                  // Running state with layered shadows
                  'border-blue-400 bg-blue-50 dark:bg-blue-950/30 shadow-md':
                    status === 'running' && !isSelected,
                  'border-blue-500 bg-blue-100 dark:bg-blue-900/40 shadow-2xl ring-2 ring-blue-500':
                    status === 'running' && isSelected,
                  // Failed state with layered shadows
                  'border-red-400 bg-red-50 dark:bg-red-950/30 shadow-md':
                    status === 'failed' && !isSelected,
                  'border-red-500 bg-red-100 dark:bg-red-900/40 shadow-2xl ring-2 ring-red-500':
                    status === 'failed' && isSelected,
                  // Pending state with layered shadows
                  'border-cool-grey-300 bg-cool-grey-50 dark:bg-dark-grey-800/50 shadow-sm':
                    status === 'pending' && !isSelected,
                  'border-cool-grey-400 bg-cool-grey-100 dark:bg-dark-grey-700/60 shadow-xl ring-2 ring-cool-grey-400':
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

              {/* Install/Build info */}
              <div className="flex flex-col items-start flex-1 min-w-0">
                <Text variant="subtext" weight="normal" className="truncate">
                  {stage.name}
                </Text>
                <Text variant="caption" theme="neutral" className="font-mono text-xs truncate max-w-full">
                  {install.installName}
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
            style={{
              marginLeft: `${visibleInstalls.length * 4}px`,
            }}
            className={cn(
              'flex items-center justify-center gap-2 px-4 py-2.5',
              'min-w-[280px] rounded-md border-2 transition-all duration-300',
              'cursor-pointer select-none',
              'border-cool-grey-300 bg-cool-grey-100 hover:bg-cool-grey-200',
              'dark:border-dark-grey-600 dark:bg-dark-grey-700 dark:hover:bg-dark-grey-600',
              'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
              'hover:shadow-md hover:-translate-y-0.5',
              'active:scale-95'
            )}
          >
            <Icon
              variant={isExpanded ? 'chevron-up' : 'chevron-down'}
              size={16}
              className={cn(
                'text-cool-grey-600 dark:text-cool-grey-400 transition-transform duration-300',
                isExpanded && 'rotate-180'
              )}
            />
            <Text variant="subtext" weight="normal" className="text-cool-grey-700 dark:text-cool-grey-300">
              {isExpanded ? 'Show Less' : `Show ${hiddenCount} More`}
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
          // Approval stage - special amber/yellow treatment
          'border-amber-400 bg-amber-50 dark:bg-amber-950/20':
            stage.status === 'pending' && stage.name === 'Pending Approval' && !isSelected,
          'border-amber-500 bg-amber-100 dark:bg-amber-900/30 shadow-lg ring-2 ring-amber-500':
            stage.status === 'pending' && stage.name === 'Pending Approval' && isSelected,
          // Regular pending state
          'border-cool-grey-300 bg-cool-grey-50 dark:bg-dark-grey-800/50':
            stage.status === 'pending' && stage.name !== 'Pending Approval' && !isSelected,
          'border-cool-grey-400 bg-cool-grey-100 dark:bg-dark-grey-700/60 shadow-lg ring-2 ring-cool-grey-400':
            stage.status === 'pending' && stage.name !== 'Pending Approval' && isSelected,
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
      {/* Current Group Overview */}
      <Card>
        <div className="flex flex-col gap-4">
          {/* Header with status */}
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <Text variant="h4" weight="strong">
                  Current Group Overview
                </Text>
                <Status status={stage.status} variant="badge" />
              </div>
              <Text variant="base" theme="neutral">
                {stage.name} - {stage.description}
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
                  {completedSteps} / {allSteps.length}
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
              Steps ({stage.steps.length})
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

// Zoom controls component
const ZoomControls = ({
  zoom,
  onZoomIn,
  onZoomOut,
  onZoomReset,
}: {
  zoom: number
  onZoomIn: () => void
  onZoomOut: () => void
  onZoomReset: () => void
}) => {
  return (
    <div className="absolute top-4 right-4 z-10 flex flex-col items-center gap-0.5 bg-white dark:bg-dark-grey-800 rounded-lg border border-cool-grey-200 dark:border-dark-grey-600 shadow-lg p-1">
      {/* Zoom level display - at the top */}
      <div className="px-2 py-1 text-center min-w-[44px]">
        <Text variant="caption" weight="strong" className="text-cool-grey-600 dark:text-cool-grey-400 tabular-nums text-xs">
          {Math.round(zoom * 100)}%
        </Text>
      </div>
      
      {/* Zoom in button */}
      <button
        onClick={onZoomIn}
        disabled={zoom >= 2}
        title="Zoom In (Ctrl/Cmd + Scroll Up)"
        className={cn(
          'w-7 h-7 flex items-center justify-center rounded transition-all duration-150',
          'bg-cool-grey-100 dark:bg-dark-grey-700',
          'hover:bg-blue-100 dark:hover:bg-blue-900/40 hover:text-blue-600 dark:hover:text-blue-400',
          'active:scale-95 active:bg-blue-200 dark:active:bg-blue-900/60',
          'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-cool-grey-100 dark:disabled:hover:bg-dark-grey-700',
          'text-cool-grey-700 dark:text-cool-grey-300',
          'border border-cool-grey-200 dark:border-dark-grey-600'
        )}
      >
        <Icon variant="Plus" size={14} />
      </button>
      
      {/* Zoom out button */}
      <button
        onClick={onZoomOut}
        disabled={zoom <= 0.5}
        title="Zoom Out (Ctrl/Cmd + Scroll Down)"
        className={cn(
          'w-7 h-7 flex items-center justify-center rounded transition-all duration-150',
          'bg-cool-grey-100 dark:bg-dark-grey-700',
          'hover:bg-blue-100 dark:hover:bg-blue-900/40 hover:text-blue-600 dark:hover:text-blue-400',
          'active:scale-95 active:bg-blue-200 dark:active:bg-blue-900/60',
          'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-cool-grey-100 dark:disabled:hover:bg-dark-grey-700',
          'text-cool-grey-700 dark:text-cool-grey-300',
          'border border-cool-grey-200 dark:border-dark-grey-600'
        )}
      >
        <Icon variant="Minus" size={14} />
      </button>
      
      {/* Divider */}
      <div className="w-5 h-px bg-cool-grey-200 dark:bg-dark-grey-600 my-0.5" />
      
      {/* Reset zoom button - resets to 100% */}
      <button
        onClick={onZoomReset}
        title="Reset to 100%"
        className={cn(
          'w-7 h-7 flex items-center justify-center rounded transition-all duration-150',
          'bg-cool-grey-100 dark:bg-dark-grey-700',
          'hover:bg-blue-100 dark:hover:bg-blue-900/40 hover:text-blue-600 dark:hover:text-blue-400',
          'active:scale-95 active:bg-blue-200 dark:active:bg-blue-900/60',
          'text-cool-grey-700 dark:text-cool-grey-300',
          'border border-cool-grey-200 dark:border-dark-grey-600'
        )}
      >
        <Icon variant="ArrowsInSimple" size={14} />
      </button>
    </div>
  )
}

// Draggable canvas component with zoom (no momentum)
const DraggableCanvas = ({
  children,
  zoom,
  onZoomChange,
}: {
  children: React.ReactNode
  zoom: number
  onZoomChange: (newZoom: number) => void
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  
  // All drag state in refs for synchronous access
  const dragStateRef = useRef({
    isDragging: false,
    startX: 0,
    startY: 0,
    scrollLeft: 0,
    scrollTop: 0,
    hasMoved: false,
  })
  
  // Visual state for cursor
  const [isDraggingVisual, setIsDraggingVisual] = useState(false)
  
  // Use refs for zoom to avoid stale closures in wheel handler
  const zoomRef = useRef(zoom)
  const onZoomChangeRef = useRef(onZoomChange)
  
  // Keep refs in sync with props
  useEffect(() => {
    zoomRef.current = zoom
    onZoomChangeRef.current = onZoomChange
  }, [zoom, onZoomChange])

  // Use native event listeners for reliable drag handling
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const handleMouseDown = (e: MouseEvent) => {
      // Always start drag on mousedown in the container
      dragStateRef.current = {
        isDragging: true,
        startX: e.pageX - container.offsetLeft,
        startY: e.pageY - container.offsetTop,
        scrollLeft: container.scrollLeft,
        scrollTop: container.scrollTop,
        hasMoved: false,
      }
      setIsDraggingVisual(true)
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!dragStateRef.current.isDragging) return
      
      e.preventDefault()
      
      const x = e.pageX - container.offsetLeft
      const y = e.pageY - container.offsetTop
      const deltaX = x - dragStateRef.current.startX
      const deltaY = y - dragStateRef.current.startY
      
      // Mark as moved if we've dragged more than 3 pixels
      if (Math.abs(deltaX) > 3 || Math.abs(deltaY) > 3) {
        dragStateRef.current.hasMoved = true
      }
      
      // Apply drag with zoom-adjusted sensitivity
      const walkX = deltaX * 1.5
      const walkY = deltaY * 1.5
      container.scrollLeft = dragStateRef.current.scrollLeft - walkX
      container.scrollTop = dragStateRef.current.scrollTop - walkY
    }

    const handleMouseUp = () => {
      dragStateRef.current.isDragging = false
      setIsDraggingVisual(false)
    }

    const handleMouseLeave = () => {
      if (dragStateRef.current.isDragging) {
        dragStateRef.current.isDragging = false
        setIsDraggingVisual(false)
      }
    }

    // Wheel handler for zoom
    const handleWheel = (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault()
        const zoomDelta = e.deltaY > 0 ? -0.1 : 0.1
        const newZoom = Math.max(0.5, Math.min(2, zoomRef.current + zoomDelta))
        onZoomChangeRef.current(newZoom)
      }
    }

    // Add event listeners
    container.addEventListener('mousedown', handleMouseDown)
    container.addEventListener('mousemove', handleMouseMove)
    container.addEventListener('mouseup', handleMouseUp)
    container.addEventListener('mouseleave', handleMouseLeave)
    container.addEventListener('wheel', handleWheel, { passive: false })

    // Cleanup
    return () => {
      container.removeEventListener('mousedown', handleMouseDown)
      container.removeEventListener('mousemove', handleMouseMove)
      container.removeEventListener('mouseup', handleMouseUp)
      container.removeEventListener('mouseleave', handleMouseLeave)
      container.removeEventListener('wheel', handleWheel)
    }
  }, [])

  // Center the canvas on mount
  useEffect(() => {
    if (containerRef.current && contentRef.current) {
      const containerWidth = containerRef.current.offsetWidth
      const containerHeight = containerRef.current.offsetHeight
      const contentWidth = contentRef.current.offsetWidth
      const contentHeight = contentRef.current.offsetHeight
      const centerPositionX = (contentWidth - containerWidth) / 2
      const centerPositionY = (contentHeight - containerHeight) / 2
      containerRef.current.scrollLeft = Math.max(0, centerPositionX)
      containerRef.current.scrollTop = Math.max(0, centerPositionY)
    }
  }, [])

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative overflow-auto select-none', // overflow-auto is REQUIRED for drag-to-scroll functionality
        'bg-cool-grey-50 dark:bg-dark-grey-900 rounded-lg border-2 border-cool-grey-200 dark:border-dark-grey-700',
        'h-[calc(100vh-450px)] min-h-[400px] w-full', // Viewport-based height with minimum, full width
        {
          'cursor-grabbing': isDraggingVisual,
          'cursor-grab': !isDraggingVisual,
        }
      )}
      style={{
        scrollbarWidth: 'none', // Firefox - hide scrollbar but keep scroll functionality
        msOverflowStyle: 'none', // IE/Edge - hide scrollbar but keep scroll functionality
      }}
    >
      {/* Hide scrollbar for Chrome/Safari */}
      <style jsx>{`
        div::-webkit-scrollbar {
          display: none;
        }
      `}</style>

      <div 
        ref={contentRef} 
        className="inline-flex items-center gap-0 min-h-full transition-transform duration-200 origin-center pointer-events-none p-12"
        style={{
          transform: `scale(${zoom})`,
          transformOrigin: 'center center',
        }}
      >
        {/* Re-enable pointer events for children so clicks work */}
        <div className="pointer-events-auto inline-flex items-center justify-center gap-0 min-h-full">
          {children}
        </div>
      </div>

    </div>
  )
}

export const BranchWorkflowCanvas = ({
  workflow,
  branchId,
  appId,
  orgId,
}: IBranchWorkflowCanvas) => {
  const [stages, setStages] = useState<IWorkflowStage[]>([])
  const [selectedStage, setSelectedStage] = useState<IWorkflowStage | null>(null)
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set())
  const [selectedStep, setSelectedStep] = useState<IWorkflowStep | null>(null)
  const [isPanelOpen, setIsPanelOpen] = useState(false)
  const [expandedParallelStages, setExpandedParallelStages] = useState<Set<string>>(new Set())
  const [zoom, setZoom] = useState(1)

  // Transform workflow to stages whenever workflow changes
  useEffect(() => {
    if (!workflow) return

    const transformedStages = transformWorkflowToStages(workflow)
    setStages(transformedStages)
    
    if (transformedStages.length > 0) {
      setSelectedStage(transformedStages[0])
    }
  }, [workflow])

  const ZOOM_LEVELS = [0.5, 0.75, 1, 1.25, 1.5, 2]

  const handleZoomIn = () => {
    const currentIndex = ZOOM_LEVELS.findIndex(level => level >= zoom)
    if (currentIndex < ZOOM_LEVELS.length - 1) {
      setZoom(ZOOM_LEVELS[currentIndex + 1])
    }
  }

  const handleZoomOut = () => {
    const currentIndex = ZOOM_LEVELS.findIndex(level => level >= zoom)
    if (currentIndex > 0) {
      setZoom(ZOOM_LEVELS[currentIndex - 1])
    }
  }

  const handleZoomReset = () => {
    setZoom(1)
  }

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

  if (!workflow) {
    return (
      <div className="flex items-center justify-center h-64">
        <Text variant="base" theme="neutral">
          No workflow data available
        </Text>
      </div>
    )
  }

  if (stages.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <Text variant="base" theme="neutral">
          No workflow execution data available
        </Text>
      </div>
    )
  }

  if (!selectedStage) {
    return (
      <div className="flex items-center justify-center h-64">
        <Text variant="base" theme="neutral">
          No stage selected
        </Text>
      </div>
    )
  }

  return (
    <div className="w-full flex flex-col gap-8 overflow-x-hidden">
      {/* Side panel for step details */}
      <StepDetailSidePanel
        step={selectedStep}
        isOpen={isPanelOpen}
        onClose={handleClosePanel}
      />

      {/* Canvas section label */}
      <div className="flex items-center gap-2">
        <Text variant="base" weight="strong">Workflow Canvas</Text>
      </div>

      {/* Draggable canvas with zoom controls */}
      <div className="relative w-full">
        <ZoomControls
          zoom={zoom}
          onZoomIn={handleZoomIn}
          onZoomOut={handleZoomOut}
          onZoomReset={handleZoomReset}
        />
        <DraggableCanvas zoom={zoom} onZoomChange={setZoom}>
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
        
        {/* Canvas navigation hints */}
        <div className="flex items-center justify-center gap-4 mt-2">
          <Text variant="subtext" theme="neutral" className="italic opacity-60">
            Drag to navigate • Ctrl/Cmd + Scroll to zoom
          </Text>
        </div>
      </div>

      {/* Detail section */}
      {selectedStage && (
        <div className="w-full">
          <StageDetailSection 
            stage={selectedStage}
            expandedSteps={expandedSteps}
            onToggleStep={handleToggleStep}
            onOpenStepPanel={handleOpenStepPanel}
          />
        </div>
      )}

    </div>
  )
}
