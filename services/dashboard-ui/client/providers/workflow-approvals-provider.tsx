import { createContext, useEffect, useRef, type ReactNode } from 'react'
import { useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { getPendingApprovals } from '@/lib'
import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useOrg } from '@/hooks/use-org'
import { useOrgStatusSSE } from '@/hooks/use-org-status-sse'
import { useNotifications } from '@/hooks/use-notifications'
import { useToast } from '@/hooks/use-toast'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { toSentenceCase } from '@/utils/string-utils'
import type { TWorkflowStepApproval } from '@/types'

type WorkflowApprovalsContextValue = {
  approvals: TWorkflowStepApproval[]
  isLoading: boolean
  refresh: () => void
}

export const WorkflowApprovalsContext = createContext<
  WorkflowApprovalsContextValue | undefined
>(undefined)

export function WorkflowApprovalsProvider({
  children,
}: {
  children: ReactNode
}) {
  const { org } = useOrg()
  const { activeWorkflows } = useActiveWorkflows()
  const { sseConnected } = useOrgStatusSSE()
  const { addToast } = useToast()
  const { emitNotification } = useNotifications()
  const navigate = useNavigate()
  const seenIds = useRef<Set<string>>(new Set())
  const initialized = useRef(false)

  const {
    data: approvals,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['workflow-approvals', org.id],
    queryFn: () => getPendingApprovals({ orgId: org.id }),
    refetchInterval: sseConnected ? false : 20_000,
  })

  useEffect(() => {
    if (!approvals) return

    if (!initialized.current) {
      for (const approval of approvals) {
        seenIds.current.add(approval.id)
      }
      initialized.current = true
      return
    }

    for (const approval of approvals) {
      if (!seenIds.current.has(approval.id)) {
        seenIds.current.add(approval.id)
        const step = approval.workflow_step
        const stepName = step?.name
        const installName = step?.owner_id
          ? activeWorkflows.find((w) => w.owner_id === step.owner_id)?.metadata
              ?.owner_name
          : undefined
        const heading = stepName ? toSentenceCase(stepName) : 'Approval required'
        const workflowUrl =
          step?.owner_id && step?.install_workflow_id
            ? `/${org.id}/installs/${step.owner_id}/workflows/${step.install_workflow_id}`
            : null
        addToast(
          <Toast
            heading={installName ? `${installName} — ${heading}` : heading}
            theme="warn"
          >
            <Text>Workflow step needs approved.</Text>
            {workflowUrl ? (
              <Link href={workflowUrl}>
                View details <Icon variant="CaretRightIcon" />
              </Link>
            ) : (
              'A workflow step is waiting for your approval.'
            )}
          </Toast>
        )
        emitNotification({
          title: installName ? `${installName} — ${heading}` : heading,
          body: 'A workflow step is waiting for your approval.',
          icon: '/favicon.svg',
          tag: approval.id,
          onClick: () => {
            window.focus()
            if (workflowUrl) {
              navigate(workflowUrl)
            }
          },
        })
      }
    }
  }, [approvals, activeWorkflows, addToast, emitNotification, navigate, org.id])

  return (
    <WorkflowApprovalsContext.Provider
      value={{ approvals: approvals ?? [], isLoading, refresh: refetch }}
    >
      {children}
    </WorkflowApprovalsContext.Provider>
  )
}
