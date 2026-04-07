import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { getApprovalType } from '@/utils/approval-utils'
import { toSentenceCase } from '@/utils/string-utils'
import type { TWorkflowStepApproval, TWorkflow } from '@/types'

interface IPendingApprovals {
  orgId: string
  approvals: TWorkflowStepApproval[]
  activeWorkflows: TWorkflow[]
}

export const PendingApprovals = ({ orgId, approvals, activeWorkflows }: IPendingApprovals) => {
  const ownerNames = new Map(
    activeWorkflows
      .filter((w) => w.owner_id && w.metadata?.owner_name)
      .map((w) => [w.owner_id!, w.metadata!.owner_name!])
  )

  if (approvals.length === 0) return null

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <Text variant="base" weight="strong">
          Pending approvals
        </Text>
        <Badge theme="warn" size="sm" variant="code">
          {approvals.length}
        </Badge>
      </div>
      <Card className="!p-0 !gap-0 overflow-hidden">
        {approvals.map((approval, i) => {
          const step = approval.workflow_step
          const href =
            step?.owner_id && step?.install_workflow_id
              ? `/${orgId}/installs/${step.owner_id}/workflows/${step.install_workflow_id}`
              : undefined
          const name = step?.name
            ? toSentenceCase(step.name)
            : 'Approval required'
          const installName = step?.owner_id
            ? ownerNames.get(step.owner_id)
            : undefined

          return (
            <div
              key={approval.id}
              className={`flex items-center justify-between px-4 py-3 gap-4 ${i < approvals.length - 1 ? 'border-b' : ''}`}
            >
              <div className="flex items-center gap-3 min-w-0">
                <Status status="pending-approval" variant="badge" />
                {href ? (
                  <Link
                    href={href}
                    className="truncate text-sm font-strong flex items-center gap-2"
                  >
                    {installName && (
                      <>
                        <Text>{installName}</Text>
                        <span>|</span>
                      </>
                    )}
                    {name}
                  </Link>
                ) : (
                  <Text className="truncate">
                    {installName && <>{installName} | </>}
                    {name}
                  </Text>
                )}
              </div>
              {approval.type && (
                <Badge theme="warn" size="sm" variant="code">
                  {getApprovalType(approval.type)}
                </Badge>
              )}
            </div>
          )
        })}
      </Card>
    </div>
  )
}
