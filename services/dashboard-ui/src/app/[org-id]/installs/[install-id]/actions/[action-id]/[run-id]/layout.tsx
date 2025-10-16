import { InstallActionRunHeader } from '@/components/actions/InstallActionRunHeader'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { TabNav } from '@/components/navigation/TabNav'
import {
  getInstallActionById,
  getInstallActionRunById,
  getWorkflowById,
  getOrgById,
} from '@/lib'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import type { TLayoutProps } from '@/types'

type TInstallActionRunLayout = TLayoutProps<
  'org-id' | 'install-id' | 'action-id' | 'run-id'
>

export default async function InstallActionRunLayout({
  children,
  params,
}: TInstallActionRunLayout) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [{ data: installActionRun }, { data: installAction }, { data: org }] =
    await Promise.all([
      getInstallActionRunById({
        installId,
        orgId,
        runId,
      }),
      getInstallActionById({
        actionId,
        installId,
        orgId,
      }),
      getOrgById({ orgId }),
    ])

  const { data: workflow } = await getWorkflowById({
    orgId,
    workflowId: installActionRun?.install_workflow_id,
  })

  const containerId = 'action-run-page'
  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun}
      shouldPoll
    >
      {org?.features?.['stratus-layout'] ? (
        <PageSection id={containerId} isScrollable>
          <InstallActionRunHeader
            actionId={actionId}
            actionName={installAction?.action_workflow?.name}
            workflow={workflow}
          />
          <TabNav
            basePath={`/${orgId}/installs/${installId}/actions/${actionId}/${runId}`}
            tabs={[
              { text: 'Summary', path: '/' },
              { text: 'Logs', path: '/logs' },
            ]}
          />
          {children}
          <BackToTop containerId={containerId} />
        </PageSection>
      ) : (
        children
      )}
    </InstallActionRunProvider>
  )
}
