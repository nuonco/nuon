import { Banner } from '@/components/common/Banner'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TInstallRoleUsage } from '@/types'

interface IInstallRoleUsagesModal extends IModal {
  orgId: string
  installId: string
  usages?: TInstallRoleUsage[]
  isLoading: boolean
  error: TAPIError | null
  roleDisplayName?: string
  onNavigate?: () => void
}

const principalFromUsage = (
  usage: TInstallRoleUsage,
): { label: string; value: string } => {
  const job = usage.runner_job
  const metadata = job?.metadata as Record<string, string> | undefined
  switch (job?.owner_type) {
    case 'install_deploys':
      return {
        label: 'Component',
        value: metadata?.component_name || '—',
      }
    case 'install_sandbox_runs':
      return {
        label: 'Sandbox',
        value: metadata?.sandbox_run_type || 'run',
      }
    case 'install_action_workflow_runs':
      return {
        label: 'Action',
        value: metadata?.action_workflow_name || '—',
      }
    default:
      return {
        label: 'Owner',
        value: job?.owner_type || '—',
      }
  }
}

export const InstallRoleUsagesModal = ({
  orgId,
  installId,
  usages,
  isLoading,
  error,
  roleDisplayName,
  onNavigate,
  ...props
}: IInstallRoleUsagesModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="ShieldCheck" size="24" />
          Role usage{roleDisplayName ? ` — ${roleDisplayName}` : ''}
        </Text>
      }
      className="!max-w-3xl !max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to load role usage.'}
          </Banner>
        ) : null}

        {isLoading ? (
          <div className="flex flex-col gap-2">
            {Array.from({ length: 3 }).map((_, idx) => (
              <Skeleton key={idx} height="60px" width="100%" />
            ))}
          </div>
        ) : usages?.length ? (
          <div className="flex flex-col divide-y">
            {usages.map((usage) => {
              const principal = principalFromUsage(usage)
              const workflowId = usage.workflow?.id
              const stepId = usage.workflow_step_id
              const workflowName =
                usage.workflow?.name || usage.workflow?.type || 'Workflow'
              const href = workflowId
                ? `/${orgId}/installs/${installId}/workflows/${workflowId}${
                    stepId ? `?panel=${stepId}` : ''
                  }`
                : null

              const operation = usage.runner_job?.operation
              const operationLabel = operation?.includes('plan')
                ? operation === 'apply-plan'
                  ? 'apply'
                  : 'plan'
                : operation

              return (
                <div
                  key={usage.id}
                  className="flex items-start justify-between gap-4 py-3"
                >
                  <div className="flex flex-col gap-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <Text variant="base" weight="strong">
                        {workflowName}
                      </Text>
                      {operationLabel ? (
                        <Badge variant="code" size="sm">
                          {operationLabel}
                        </Badge>
                      ) : null}
                    </div>
                    <Text variant="subtext" theme="neutral">
                      {principal.label}: {principal.value}
                      {usage.runner_job?.type
                        ? ` · ${usage.runner_job.type}`
                        : ''}
                    </Text>
                    {usage.runner_job?.created_at ? (
                      <Time
                        variant="subtext"
                        time={usage.runner_job.created_at}
                        format="long-datetime"
                      />
                    ) : null}
                  </div>
                  {href ? (
                    <Link href={href} variant="default" onClick={onNavigate}>
                      View workflow
                      <Icon variant="CaretRight" size="16" />
                    </Link>
                  ) : null}
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex items-center justify-center p-8">
            <Text variant="body" theme="neutral">
              This role has not been used by any workflow yet.
            </Text>
          </div>
        )}
      </div>
    </Modal>
  )
}

interface IInstallRoleUsagesTrigger {
  onOpenModal: () => void
}

export const InstallRoleUsagesTrigger = ({
  onOpenModal,
}: IInstallRoleUsagesTrigger) => {
  return (
    <Link
      href="#"
      variant="default"
      className="text-sm"
      onClick={(e) => {
        e.preventDefault()
        onOpenModal()
      }}
    >
      View usage
      <Icon variant="CaretRight" size="14" />
    </Link>
  )
}
