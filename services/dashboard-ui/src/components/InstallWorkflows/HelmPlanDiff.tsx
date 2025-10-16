'use client'

import React, { useState } from 'react'
import { CaretRightIcon } from '@phosphor-icons/react'
import { Badge } from "../Badge"
import { CodeBlock } from '../CodeBlock'
import { Text } from '../Typography'
import { diffEntries, diffLines } from './diff-utils'

interface Change {
  namespace: string
  release: string
  resource: string
  resourceType: string
  action: string
}

interface PlanSummary {
  add: number
  change: number
  destroy: number
}

interface HelmPlan {
  op: string
  plan: string
  helm_content_diff?: HelmContentDiff[]
  k8s_content_diff?: HelmContentDiff[]
}

// payload contains n lines seprated by \n
// delta 1 : before
// delta 2 : after
export interface HelmContentDiffEntry {
  delta?: 1 | 2 | 0;
  type?: 1 | 2 | 0;  // Added type field which is an alternative to delta
  payload?: string;
}

interface HelmContentDiff {
  _version: string
  api: string
  name: string
  namespace: string
  kind: string
  before: string // YAML string
  after: string // YAML string
  entries: HelmContentDiffEntry[]
}

interface HelmChangesViewerProps {
  planData: HelmPlan
}

/**
 * Correct matching function based on your data sample.
 */
function findDiffForChange(change: Change, diffs?: HelmContentDiff[]) {
  if (!diffs) return undefined
  return diffs.find(
    (d) =>
      d.api === change.resourceType &&
      d.kind === change.resource &&
      d.namespace === change.namespace &&
      d.name === change.release
  )
}

const parseChanges = (
  planText: string
): { changes: Change[]; summary: PlanSummary } => {
  const changes: Change[] = []
  const summary: PlanSummary = { add: 0, change: 0, destroy: 0 }
  const lines = planText.replace(/\u001b\[\d+;?\d*m/g, '').split('\n')
  lines.forEach((line) => {
    const match = line.match(
      /^([^,]+),\s*([^,]+),\s*([^(]+)\s*\(([^)]+)\)\s*to\s*be\s*(\w+)/
    )
    if (match) {
      changes.push({
        namespace: match[1].trim(),
        release: match[2].trim(),
        resource: match[3].trim(),
        resourceType: match[4].trim(),
        action: match[5].trim(),
      })
    }
    const summaryMatch = line.match(
      /Plan: (\d+) to add, (\d+) to change, (\d+) to destroy/
    )
    if (summaryMatch) {
      summary.add = parseInt(summaryMatch[1])
      summary.change = parseInt(summaryMatch[2])
      summary.destroy = parseInt(summaryMatch[3])
    }
  })
  return { changes, summary }
}

export const HelmChangesViewer: React.FC<HelmChangesViewerProps> = ({
  planData,
}) => {
  const { changes, summary } = parseChanges(planData.plan)
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null)

  return (
    <div className="bg-cool-grey-50 dark:bg-dark-grey-200 rounded-lg border">
      {/* Header */}
      <div className="flex flex-col px-4 py-4 sm:px-6 border-b">
        <Text variant="med-18">
          {planData?.k8s_content_diff ? "Kubernetes" : "Helm"} changes overview
        </Text>
        <Text isMuted>Operation: {planData.op}</Text>
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
            <Text
              variant="med-14"
              className="text-red-600 dark:text-red-400"
            >
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
          const diff = findDiffForChange(change, planData.helm_content_diff || planData?.k8s_content_diff)
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
                    <Text className="font-medium">{change.release}</Text>
                    <Text variant="reg-12" isMuted>
                      {change.resource} ({change.resourceType})
                    </Text>
                    <Text variant="reg-12" isMuted>
                      Namespace: {change.namespace}
                    </Text>
                  </div>
                </span>
                <div className="flex items-center">
                  <Badge
                    theme={
                      change.action === 'changed'
                        ? 'warn'
                        : change.action === 'added'
                          ? 'success'
                          : 'error'
                    }
                  >
                    {change.action}
                  </Badge>
                </div>
              </button>

              {/* Diff or no diff when expanded */}
              {isExpanded && (
                <div className="mt-4">
                  {diff ? (
                    <div className="flex flex-col md:flex-row gap-4 px-2 py-2 bg-white dark:bg-dark-grey-100 rounded border">
                      <CodeBlock className="w-full" language="yaml" isDiff>
                        { diff?._version == '2' ? diffEntries(diff?.entries) : diffLines(diff?.before, diff?.after)}
                      </CodeBlock>
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
