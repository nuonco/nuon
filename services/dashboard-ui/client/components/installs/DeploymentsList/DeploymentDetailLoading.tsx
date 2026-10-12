import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { DEPLOYMENT_TABS } from '@/components/installs/DeploymentDetail/deployment-progress'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'

export const DeploymentDetailLoading = ({
  activity,
  createdAt,
  id,
  status,
}: {
  activity?: string
  createdAt?: string
  id?: string
  status?: string
}) => (
  <>
    <div className="flex flex-wrap items-center gap-3">
      {status ? (
        <Status status={status} variant="badge" />
      ) : (
        <Status loading variant="badge" loadingWidth={8} />
      )}
      <ID loading={!id} loadingWidth={16}>
        {id}
      </ID>
      {createdAt ? (
        <Time time={createdAt} format="relative" variant="subtext" />
      ) : (
        <Time loading format="relative" variant="subtext" loadingWidth={10} />
      )}
    </div>
    {activity ? (
      <Text variant="subtext" theme={status === 'error' ? 'error' : 'neutral'}>
        {activity}
      </Text>
    ) : (
      <Text variant="subtext" loading loadingWidth={42} />
    )}
    <Tabs
      naturalHeight
      initActiveTab="workflow"
      tabsClassName="mt-4"
      tabControlsClassName="!gap-2 md:!gap-6 [&>button]:px-1 md:[&>button]:px-3"
      tabLabels={Object.fromEntries(
        Object.entries(DEPLOYMENT_TABS).map(([key, tab]) => [key, tab.text])
      )}
      tabs={{
        template: null,
        workflow: (
          <div className="flex flex-col gap-4">
            <div className="flex justify-end">
              <Text loading loadingWidth={14} />
            </div>
            <div className="flex flex-col gap-1.5">
              <SectionHeader title="Workflow steps" />
              <Text
                variant="subtext"
                theme="neutral"
                loading
                loadingWidth={28}
              />
            </div>
            <WorkflowStepsSkeleton />
          </div>
        ),
        changes: null,
      }}
    />
  </>
)
