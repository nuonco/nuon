'use client'

import React, { useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Badge, Text } from '@/stratus/components'
import { Code } from "../Typography"

interface Change {
  workspace: string
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
  helm_content_diff?: HelmContentDiffEntry[]
}

interface HelmContentDiffEntry {
  api: string
  name: string
  namespace: string
  kind: string
  before: string // YAML string
  after: string  // YAML string
}

interface HelmChangesViewerProps {
  planData: HelmPlan
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
        workspace: match[1].trim(),
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
  const hasContentDiff = planData.helm_content_diff && planData.helm_content_diff.length > 0

  return (
    <div className="bg-cool-grey-50 dark:bg-dark-grey-200 rounded-lg border">
      {/* Header */}
      <div className="flex flex-col px-4 py-4 sm:px-6 border-b">
        <Text variant="h3" weight="strong">
          Helm Changes Overview
        </Text>
        <Text theme="muted">Operation: {planData.op}</Text>
      </div>

      {/* Summary */}
      <div className="px-4 py-3 border-b bg-cool-grey-100 dark:bg-dark-grey-300">
        <div className="flex space-x-4">
          <div className="flex items-center gap-1.5">
            <Text
              variant="base"
              className="text-green-600 dark:text-green-40"
              weight="strong"
            >
              {summary.add}
            </Text>
            <Text variant="subtext" theme="muted">
              to add
            </Text>
          </div>
          <div className="flex items-center gap-1.5">
            <Text
              variant="base"
              className="text-orange-600 dark:text-orange-400"
              weight="strong"
            >
              {summary.change}
            </Text>
            <Text variant="subtext" theme="muted">
              to change
            </Text>
          </div>
          <div className="flex items-center gap-1.5">
            <Text
              variant="base"
              className="text-red-600 dark:text-red-400"
              weight="strong"
            >
              {summary.destroy}
            </Text>
            <Text variant="subtext" theme="muted">
              to destroy
            </Text>
          </div>
        </div>
      </div>

      {/* Changes List */}
      <div className="divide-y">
        {changes.map((change, index) => (
          <div key={index} className="px-4 py-4 sm:px-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="flex flex-col">
                  <Text weight="strong">{change.release}</Text>
                  <Text variant="subtext" theme="muted">
                    {change.resource} ({change.resourceType})
                  </Text>
                </div>
              </div>
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
            </div>
            <div>
              <Text variant="subtext" theme="muted">
                Workspace: {change.workspace}
              </Text>
            </div>
          </div>
        ))}
      </div>

      {/* Helm Content Diff Section */}
      {hasContentDiff && (
        <div className="border-t bg-cool-grey-50 dark:bg-dark-grey-200">
          <div className="px-4 py-4 sm:px-6 flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Resource Diffs
            </Text>
            <div className="flex flex-col gap-2">
              {planData.helm_content_diff!.map((diff, i) => {
                const isExpanded = expandedIndex === i
                return (
                  <div key={`${diff.kind}-${diff.name}-${diff.namespace}`}>
                    <button
                      className="w-full flex items-center justify-between px-3 py-2 bg-cool-grey-100 dark:bg-dark-grey-300 transition-all rounded hover:bg-cool-grey-200 dark:hover:bg-dark-grey-100 focus:outline-none"
                      onClick={() => setExpandedIndex(isExpanded ? null : i)}
                      aria-expanded={isExpanded}
                    >
                      <span className="flex items-center gap-3">
                        <CaretRight
                          size={16}
                          className={`transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                        />
                        <Text weight="strong">{diff.kind}</Text>
                        <Text family="mono" theme="muted">{diff.name}</Text>
                        <Badge variant="code" theme="neutral">
                          ns: {diff.namespace}
                        </Badge>
                      </span>
                      <span className="ml-auto flex gap-1 text-xs text-cool-grey-600 dark:text-cool-grey-60">
                        {diff.api && (
                          <span>
                            <Badge variant="code" theme="neutral">{diff.api}</Badge>
                          </span>
                        )}
                      </span>
                    </button>
                    {isExpanded && (
                      <div className="flex flex-col md:flex-row gap-4 mt-2 px-2 py-2 bg-white dark:bg-dark-grey-100 rounded border">
                        <div className="w-full md:w-1/2 flex flex-col gap-2">
                          <Text variant="subtext" theme="muted">
                            Before
                          </Text>
                          <Code variant="preformated">
                            {diff.before || <span className="text-cool-grey-500 italic">(new resource)</span>}
                          </Code>
                        </div>
                        <div className="w-full md:w-1/2 flex flex-col gap-2">
                          <Text variant="subtext" theme="muted">
                            After
                          </Text>
                          <Code variant="preformated">
                            {diff.after}
                          </Code>
                        </div>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
