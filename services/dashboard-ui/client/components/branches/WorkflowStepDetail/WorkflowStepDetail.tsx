import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import type { TInstallWorkflowStep } from '@/types'

function statusTheme(status?: string) {
  if (status === 'success') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

function miniStatusIcon(status?: string) {
  if (status === 'success') return 'CheckCircleIcon'
  if (status === 'error') return 'XCircleIcon'
  if (status === 'in-progress') return 'CircleNotchIcon'
  if (status === 'skipped') return 'MinusCircleIcon'
  return 'ClockIcon'
}

function miniStatusColor(status?: string) {
  if (status === 'success') return 'text-green-500'
  if (status === 'error') return 'text-red-500'
  if (status === 'in-progress') return 'text-blue-500 animate-spin'
  if (status === 'skipped') return 'text-cool-grey-400'
  return 'text-cool-grey-400'
}

interface IWorkflowStepDetail {
  step: TInstallWorkflowStep
  onClose: () => void
}

export const WorkflowStepDetail = ({ step, onClose }: IWorkflowStepDetail) => {
  const metadata = step.status?.metadata || {}

  // Detect step type from name
  const isCommitStep = step.name?.toLowerCase().includes('commit')
  const isBuildStep = step.name?.toLowerCase().includes('build')
  const isConfigStep = step.name?.toLowerCase().includes('config') && !step.name?.toLowerCase().includes('diff')

  // Build data from metadata (set by builds signal)
  const builds = (metadata.builds as any[]) || []

  // Commit data from metadata (set by fetchcommit signal)
  const commitSha = metadata.commit_sha as string | undefined
  const commitMessage = metadata.commit_message as string | undefined
  const authorName = metadata.author_name as string | undefined
  const authorEmail = metadata.author_email as string | undefined

  return (
    <Card>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <Text variant="h3" weight="strong">
              {step.name || 'Step details'}
            </Text>
            <Badge theme={statusTheme(step.status?.status)}>
              {step.status?.status || 'pending'}
            </Badge>
          </div>
          <div className="flex items-center gap-3">
            {step.execution_time ? (
              <Text variant="subtext" theme="neutral" family="mono">
                {(step.execution_time / 1000000000).toFixed(1)}s
              </Text>
            ) : null}
            {step.started_at && (
              <Text variant="subtext" theme="neutral">
                <Time time={step.started_at} format="relative" />
              </Text>
            )}
            <Button variant="ghost" size="sm" onClick={onClose}>
              <Icon variant="XIcon" size={20} />
            </Button>
          </div>
        </div>

        {/* Status description */}
        {step.status?.status_human_description && !isCommitStep && !isBuildStep && (
          <div className="p-3 bg-cool-grey-100 dark:bg-dark-grey-800 rounded-md mb-4">
            <Text variant="base">
              {step.status.status_human_description}
            </Text>
          </div>
        )}

        {/* ===== COMMIT STEP: show commit details ===== */}
        {isCommitStep && commitSha && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-850 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700 mb-4">
            <div className="flex items-start gap-3">
              <Icon variant="GitCommitIcon" size={20} className="text-cool-grey-500 dark:text-dark-grey-300 mt-0.5 shrink-0" />
              <div className="min-w-0 flex-1">
                <Text variant="base" weight="strong">
                  {commitMessage?.split('\n')[0] || 'No message'}
                </Text>
                {commitMessage?.includes('\n') && (
                  <Text variant="subtext" theme="neutral" className="mt-1 whitespace-pre-wrap">
                    {commitMessage.split('\n').slice(1).join('\n').trim()}
                  </Text>
                )}
                <div className="flex items-center gap-3 mt-2">
                  <Text variant="subtext" theme="neutral" family="mono">
                    {commitSha.substring(0, 12)}
                  </Text>
                  {authorName && (
                    <Text variant="subtext" theme="neutral">
                      {authorName}
                    </Text>
                  )}
                  {authorEmail && (
                    <Text variant="subtext" theme="neutral">
                      &lt;{authorEmail}&gt;
                    </Text>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {isCommitStep && !commitSha && step.status?.status === 'pending' && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-850 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700 mb-4">
            <Text variant="subtext" theme="neutral">
              Waiting to fetch commit from VCS...
            </Text>
          </div>
        )}

        {/* ===== CONFIG STEP: show config summary ===== */}
        {isConfigStep && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-850 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700 mb-4">
            <Text variant="base" theme="neutral">
              {step.status?.status === 'success'
                ? 'Cloned repository, parsed configuration, and synced components.'
                : step.status?.status === 'in-progress'
                ? 'Cloning repository and parsing configuration...'
                : 'Waiting to fetch app configuration...'}
            </Text>
          </div>
        )}

        {/* ===== BUILD STEP: show individual builds ===== */}
        {isBuildStep && builds.length > 0 && (
          <div className="mb-4">
            <div className="flex items-center justify-between mb-2">
              <Text variant="label" theme="neutral">
                Component builds
              </Text>
              <div className="flex items-center gap-2">
                {builds.filter((b: any) => b.status === 'success').length > 0 && (
                  <Badge theme="success" size="sm">
                    {builds.filter((b: any) => b.status === 'success').length} passed
                  </Badge>
                )}
                {builds.filter((b: any) => b.status === 'error').length > 0 && (
                  <Badge theme="error" size="sm">
                    {builds.filter((b: any) => b.status === 'error').length} failed
                  </Badge>
                )}
                {builds.filter((b: any) => b.status === 'in-progress').length > 0 && (
                  <Badge theme="info" size="sm">
                    {builds.filter((b: any) => b.status === 'in-progress').length} running
                  </Badge>
                )}
                {builds.filter((b: any) => b.status === 'skipped').length > 0 && (
                  <Badge theme="neutral" size="sm">
                    {builds.filter((b: any) => b.status === 'skipped').length} skipped
                  </Badge>
                )}
              </div>
            </div>
            <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg divide-y divide-cool-grey-200 dark:divide-dark-grey-700 overflow-hidden">
              {builds.map((build: any, i: number) => (
                <div
                  key={build.component_id || i}
                  className="flex items-center justify-between px-4 py-3"
                >
                  <div className="flex items-center gap-3">
                    <Icon
                      variant={miniStatusIcon(build.status) as any}
                      size={16}
                      className={miniStatusColor(build.status)}
                    />
                    <Text variant="base">
                      {build.component_name || build.component_id}
                    </Text>
                  </div>
                  <Badge theme={statusTheme(build.status)} size="sm">
                    {build.status || 'pending'}
                  </Badge>
                </div>
              ))}
            </div>
          </div>
        )}

        {isBuildStep && builds.length === 0 && step.status?.status === 'pending' && (
          <div className="p-4 bg-cool-grey-50 dark:bg-dark-grey-850 rounded-lg border border-cool-grey-200 dark:border-dark-grey-700 mb-4">
            <Text variant="subtext" theme="neutral">
              Waiting to start component builds...
            </Text>
          </div>
        )}

        {/* Footer metadata */}
        <div className="flex items-center gap-4 pt-2 border-t border-cool-grey-200 dark:border-dark-grey-700">
          <ID>{step.id}</ID>
          {step.install_workflow_id && (
            <AdminDashboardLink
              path={`/workflows/${step.install_workflow_id}`}
              label="Admin panel"
            />
          )}
        </div>
      </div>
    </Card>
  )
}
