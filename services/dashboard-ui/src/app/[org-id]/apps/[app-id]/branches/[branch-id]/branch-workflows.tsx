'use client'

import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Time } from '@/components/common/Time'
import { Badge } from '@/components/common/Badge'
import type { TAppBranch } from '@/types'

interface IBranchWorkflows {
  branch: TAppBranch
}

export const BranchWorkflows = ({ branch }: IBranchWorkflows) => {
  const workflows = branch.workflows || []

  if (workflows.length === 0) {
    return (
      <EmptyState
        emptyTitle="No workflows yet"
        emptyMessage="Workflows will appear here once the branch starts processing updates."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {workflows.map((workflow) => (
        <div
          key={workflow.id}
          className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
        >
          <div className="flex items-start justify-between">
            <div className="flex flex-col gap-2">
              <div className="flex items-center gap-2">
                <span className="font-medium">{workflow.type}</span>
                {workflow.status?.status && (
                  <Badge
                    theme={
                      workflow.status.status === 'success'
                        ? 'success'
                        : workflow.status.status === 'error'
                          ? 'error'
                          : workflow.status.status === 'in-progress' ||
                              workflow.status.status === 'queued'
                            ? 'info'
                            : 'default'
                    }
                  >
                    {workflow.status.status}
                  </Badge>
                )}
              </div>

              {workflow.id && (
                <code className="text-xs text-gray-500">{workflow.id}</code>
              )}

              <div className="flex items-center gap-4 text-sm text-gray-600">
                {workflow.created_at && (
                  <span>
                    Started <Time time={workflow.created_at} format="relative" />
                  </span>
                )}
                {workflow.finished_at && (
                  <span>
                    Finished <Time time={workflow.finished_at} format="relative" />
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}