'use client'

import React, { useState } from 'react'
import { CaretRightIcon } from '@phosphor-icons/react'
import { Badge } from '../Badge'
import { Text } from '../Typography'
import { DiffBlock } from './DiffBlock'
import { formatK8SDiff, K8SDiffEntry } from './diff-utils'

// Types for K8S content diff
type K8SContentDiffEntry = K8SDiffEntry

interface K8SContentDiff {
  _version: string
  name: string
  namespace: string
  kind: string
  api: string
  resource: string
  op: string
  type: number // 0 = unchanged, 1 = deleted, 2 = created, 3 = modified
  dry_run: boolean
  entries: K8SContentDiffEntry[]
}

interface K8SPlan {
  op: string
  plan: string
  k8s_content_diff?: K8SContentDiff[]
}

interface K8SChange {
  name: string
  namespace: string
  kind: string
  resource: string
  action: string
}

interface PlanSummary {
  add: number
  change: number
  destroy: number
}

interface K8SPlanDiffProps {
  planData: K8SPlan
}

/**
 * Translates K8S operation types to human-readable actions
 */
const getActionFromType = (type: number): string => {
  switch (type) {
    case 0:
      return 'unchanged'
    case 1:
      return 'deleted'
    case 2:
      return 'created'
    case 3:
      return 'modified'
    default:
      return 'unknown'
  }
}

/**
 * Try to parse JSON string if it's a string
 */
function tryParseJSON(jsonString: any): any {
  if (typeof jsonString !== 'string') {
    return jsonString
  }

  try {
    return JSON.parse(jsonString)
  } catch (e) {
    console.error('Failed to parse JSON string:', e)
    return null
  }
}

/**
 * Parses K8S content diff into a format we can use
 */
const parseK8SChanges = (
  planData: K8SPlan
): { changes: K8SChange[]; summary: PlanSummary } => {
  const changes: K8SChange[] = []
  const summary: PlanSummary = { add: 0, change: 0, destroy: 0 }

  let k8sDiffs = planData.k8s_content_diff

  // If k8s_content_diff is not available, try to extract it from plan string
  if (!k8sDiffs && planData.plan) {
    const parsedPlan = tryParseJSON(planData.plan)
    if (parsedPlan && parsedPlan.k8s_content_diff) {
      k8sDiffs = parsedPlan.k8s_content_diff
    }
  }

  if (!k8sDiffs || k8sDiffs.length === 0) {
    return { changes, summary }
  }

  // Process each diff entry
  k8sDiffs.forEach((diff) => {
    const action = getActionFromType(diff.type)

    // Add to appropriate summary counter
    if (diff.type === 2) summary.add++
    else if (diff.type === 3) summary.change++
    else if (diff.type === 1) summary.destroy++

    changes.push({
      name: diff.name,
      namespace: diff.namespace,
      kind: diff.kind,
      resource: diff.resource,
      action: action,
    })
  })

  return { changes, summary }
}

export const K8SPlanDiff: React.FC<K8SPlanDiffProps> = ({ planData }) => {
  const [parsedPlanData, setParsedPlanData] = useState<K8SPlan>(() => {
    // If the plan field is a string that contains JSON, parse it
    if (typeof planData.plan === 'string' && !planData.k8s_content_diff) {
      try {
        const parsed = JSON.parse(planData.plan)
        if (parsed.k8s_content_diff) {
          return {
            ...planData,
            k8s_content_diff: parsed.k8s_content_diff,
          }
        }
      } catch (e) {
        console.error('Failed to parse plan data:', e)
      }
    }
    return planData
  })

  const { changes, summary } = parseK8SChanges(parsedPlanData)
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null)

  // If no changes, return early with a message
  if (changes.length === 0) {
    return (
      <div className="bg-cool-grey-50 dark:bg-dark-grey-200 rounded-lg border p-4">
        <Text variant="med-18">Kubernetes changes overview</Text>
        <div className="mt-4 p-4 bg-white dark:bg-dark-grey-100 rounded border">
          <Text>No Kubernetes changes detected.</Text>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-cool-grey-50 dark:bg-dark-grey-200 rounded-lg border">
      {/* Header */}
      <div className="flex flex-col px-4 py-4 sm:px-6 border-b">
        <Text variant="med-18">Kubernetes changes overview</Text>
        <Text isMuted>Operation: {parsedPlanData.op}</Text>
      </div>

      {/* Summary */}
      <div className="px-4 py-3 border-b bg-cool-grey-100 dark:bg-dark-grey-300">
        <div className="flex space-x-4">
          <div className="flex items-center gap-1.5">
            <Text
              variant="reg-14"
              className="text-green-600 dark:text-green-40 font-medium"
            >
              {summary.add}
            </Text>
            <Text variant="reg-12" isMuted>
              to add
            </Text>
          </div>
          <div className="flex items-center gap-1.5">
            <Text
              variant="reg-14"
              className="text-orange-600 dark:text-orange-400 font-medium"
            >
              {summary.change}
            </Text>
            <Text variant="reg-12" isMuted>
              to change
            </Text>
          </div>
          <div className="flex items-center gap-1.5">
            <Text variant="med-14" className="text-red-600 dark:text-red-400">
              {summary.destroy}
            </Text>
            <Text variant="reg-12" isMuted>
              to destroy
            </Text>
          </div>
        </div>
      </div>

      {/* Changes List */}
      <div className="divide-y">
        {changes.map((change, index) => {
          const isExpanded = expandedIndex === index
          const diffEntries = parsedPlanData.k8s_content_diff?.[index]?.entries

          return (
            <div key={index} className="px-4 py-4 sm:px-6">
              {/* Change row */}
              <button
                type="button"
                className="w-full flex items-center justify-between gap-3 bg-transparent border-none focus:outline-none px-2 !shadow-none"
                onClick={() => setExpandedIndex(isExpanded ? null : index)}
                aria-expanded={isExpanded}
              >
                <span className="flex items-center gap-2">
                  <CaretRightIcon
                    size={16}
                    className={`transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                  />
                  <div className="flex flex-col text-left">
                    <Text className="font-medium">{change.name}</Text>
                    <Text variant="reg-12" isMuted>
                      {change.kind} ({change.resource})
                    </Text>
                    <Text variant="reg-12" isMuted>
                      Namespace: {change.namespace}
                    </Text>
                  </div>
                </span>
                <div className="flex items-center">
                  <Badge
                    theme={
                      change.action === 'modified'
                        ? 'warn'
                        : change.action === 'created'
                          ? 'success'
                          : 'error'
                    }
                  >
                    {change.action}
                  </Badge>
                </div>
              </button>

              {/* Diff content when expanded */}
              {isExpanded && (
                <div className="mt-4">
                  {diffEntries ? (
                    <div className="flex flex-col md:flex-row gap-4 px-2 py-2 bg-white dark:bg-dark-grey-100 rounded border">
                      <DiffBlock className="w-full">
                        {formatK8SDiff(diffEntries)}
                      </DiffBlock>
                    </div>
                  ) : (
                    <div className="mt-2 px-2 py-2 bg-cool-grey-100 dark:bg-dark-grey-200 rounded border">
                      <Text variant="reg-12" isMuted>
                        No diff available for this change.
                      </Text>
                    </div>
                  )}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
