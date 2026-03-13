import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallTopology } from '@/components/installs/overview/InstallTopology'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'
import { snakeToWords, toSentenceCase } from '@/utils/string-utils'

function MetadataBar() {
  const { install } = useInstall()

  const driftCount = install?.drifted_objects?.length ?? 0
  const cloudPlatform = install?.cloud_platform ?? ''
  const region =
    install?.aws_account?.region ??
    install?.azure_account?.location ??
    ''
  const cloudLabel = [cloudPlatform.toUpperCase(), region]
    .filter(Boolean)
    .join(' ')

  const sep = <div className="w-px h-3.5 bg-cool-grey-300 dark:bg-dark-grey-500 shrink-0" />

  return (
    <div className="flex items-center gap-3 px-4 py-2.5 border-b">
      {cloudLabel && (
        <>
          <div className="flex items-center gap-1.5">
            <Icon variant="Cloud" size={16} weight="bold" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
            <Text variant="body" theme="neutral">{cloudLabel}</Text>
          </div>
          {sep}
        </>
      )}

      <Link href="workflows" className="no-underline">
        {driftCount > 0 ? (
          <div className="flex items-center gap-1.5">
            <Icon variant="WarningCircle" size={16} weight="bold" className="text-orange-500 shrink-0" />
            <Text variant="body" theme="warn">
              {driftCount} drifted {driftCount === 1 ? 'object' : 'objects'}
            </Text>
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <Icon variant="CheckCircle" size={16} weight="bold" className="text-green-600 dark:text-green-500 shrink-0" />
            <Text variant="body" theme="neutral">No drift</Text>
          </div>
        )}
      </Link>

      {install?.runner_status && (
        <>
          {sep}
          <Link href="runner" className="no-underline">
            <div className="flex items-center gap-1.5">
              <Icon variant="SneakerMove" size={16} weight="bold" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
              <Text variant="body" theme="neutral">Runner</Text>
              <Status status={install.runner_status} variant="badge" />
            </div>
          </Link>
        </>
      )}

      {install?.updated_at && (
        <div className="ml-auto flex items-center gap-1.5">
          <Icon variant="Clock" size={16} weight="bold" className="text-cool-grey-500 dark:text-cool-grey-400 shrink-0" />
          <Text variant="body" theme="neutral">
            <Time time={install.updated_at} />
          </Text>
        </div>
      )}
    </div>
  )
}

function ActiveWorkflowBar() {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: workflowsResult } = useQuery({
    queryKey: ['install-active-workflow', org?.id, install?.id],
    queryFn: () =>
      getInstallWorkflows({
        installId: install.id,
        orgId: org.id,
        limit: 1,
        offset: 0,
        planonly: false,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 10_000,
  })

  const workflow = workflowsResult?.data?.[0]
  if (!workflow || workflow.finished) return null

  const steps = workflow.steps ?? []
  // Find the step that is currently in-flight: started but not yet finished
  const activeStep =
    steps.find((s) => s.started_at && !s.finished) ??
    steps.find(
      (s) =>
        s.status?.status &&
        ['in-progress', 'provisioning', 'building', 'applying', 'planning', 'checking-plan', 'generating', 'retrying', 'deleting', 'active'].includes(
          s.status.status,
        ),
    )
  const completedCount = steps.filter((s) => s.finished).length
  const totalCount = steps.length
  const workflowLabel = toSentenceCase(snakeToWords(workflow.type ?? ''))

  const stepLabel = activeStep?.name
    ? toSentenceCase(activeStep.name.replace(/_/g, ' '))
    : null

  return (
    <Link href={`workflows/${workflow.id}`} className="no-underline block !w-full">
      <div className="flex items-center justify-between gap-4 border-t px-4 py-3 bg-cool-grey-100 dark:bg-dark-grey-800 hover:bg-cool-grey-200 dark:hover:bg-dark-grey-700 transition-colors">
        <div className="flex items-center gap-3 min-w-0">
          <svg
            className="animate-spin h-4 w-4 shrink-0 text-cool-grey-500 dark:text-cool-grey-400"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle className="opacity-25" cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="2" />
            <path className="opacity-75" stroke="currentColor" strokeWidth="2" strokeLinecap="round" d="M4 12a8 8 0 018-8" />
          </svg>
          <Text variant="body" weight="strong" theme="default">
            {workflowLabel}
          </Text>
          {stepLabel && (
            <Text variant="body" theme="neutral" className="truncate">
              {stepLabel}
            </Text>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Badge theme="neutral" size="sm">
            {completedCount}/{totalCount} steps
          </Badge>
          <Icon variant="CaretRight" size={14} className="text-cool-grey-500 dark:text-cool-grey-400" />
        </div>
      </div>
    </Link>
  )
}

export const Overview = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection className="!pt-0" isScrollable>
      <PageTitle title={`Overview | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
        ]}
      />

      <div className="p-6">
        <div className="w-full border rounded-lg overflow-hidden">
          <MetadataBar />
          <InstallTopology />
          <ActiveWorkflowBar />
        </div>
      </div>
    </PageSection>
  )
}
