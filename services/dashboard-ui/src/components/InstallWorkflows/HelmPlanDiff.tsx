'use client'

import React from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Badge, Text } from '@/stratus/components'

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
}

interface HelmChangesViewerProps {
  planData: HelmPlan
}

const parseChanges = (
  planText: string
): { changes: Change[]; summary: PlanSummary } => {
  const changes: Change[] = []
  const summary: PlanSummary = { add: 0, change: 0, destroy: 0 }

  // Remove ANSI color codes and split into lines
  const lines = planText.replace(/\u001b\[\d+;?\d*m/g, '').split('\n')

  // Parse each change line
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

    // Parse summary line
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
    </div>
  )
}
