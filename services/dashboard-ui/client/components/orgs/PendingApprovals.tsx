import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import { getApprovalType } from '@/utils/approval-utils'
import { toSentenceCase } from '@/utils/string-utils'

export const PendingApprovals = () => {
  const { org } = useOrg()
  const { approvals } = useWorkflowApprovals()

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
              ? `/${org.id}/installs/${step.owner_id}/workflows/${step.install_workflow_id}`
              : undefined
          const name = step?.name
            ? toSentenceCase(step.name)
            : 'Approval required'

          return (
            <div
              key={approval.id}
              className={`flex items-center justify-between px-4 py-3 gap-4 ${i < approvals.length - 1 ? 'border-b' : ''}`}
            >
              <div className="flex items-center gap-3 min-w-0">
                <Status status="pending-approval" variant="badge" />
                {href ? (
                  <Link href={href} className="truncate text-sm font-strong">
                    {name}
                  </Link>
                ) : (
                  <Text className="truncate">{name}</Text>
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
