import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  ActionTriggerButton,
  ActionTriggerType,
  Badge,
  ClickToCopy,
  CodeViewer,
  Config,
  ConfigurationVCS,
  ConfigurationVariables,
  ErrorFallback,
  Expand,
  InstallWorkflowRunHistory,
  Loading,
  DashboardContent,
  Pagination,
  Section,
  StatusBadge,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import { getInstall, getInstallActionWorkflowRecentRun } from '@/lib'
import type { TInstallActionWorkflow } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionWorkflowId,
  } = await params
  const [install, action] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallActionWorkflowRecentRun({ actionWorkflowId, installId, orgId }),
  ])

  return {
    title: `${install.name} | ${action.action_workflow?.name}`,
  }
}

export default async function InstallWorkflowRuns({ params, searchParams }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionWorkflowId,
  } = await params
  const sp = await searchParams
  const [install, actionWithRecentRuns] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallActionWorkflowRecentRun({ actionWorkflowId, installId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/actions`,
          text: 'Actions',
        },
        {
          href: `/${orgId}/installs/${install.id}/actions/${actionWorkflowId}`,
          text: actionWithRecentRuns?.action_workflow?.name,
        },
      ]}
      heading={actionWithRecentRuns.action_workflow?.name}
      headingUnderline={actionWithRecentRuns.action_workflow?.id}
      statues={
        <div className="flex gap-6 items-start justify-start">
          {actionWithRecentRuns?.runs?.length ? (
            <>
              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Recent status
                </Text>
                <StatusBadge
                  description={
                    actionWithRecentRuns.runs?.[0]?.status_v2
                      ?.status_human_description ||
                    actionWithRecentRuns?.runs?.[0]?.status_description
                  }
                  status={
                    actionWithRecentRuns.runs?.[0]?.status_v2?.status ||
                    actionWithRecentRuns?.runs?.[0]?.status ||
                    'noop'
                  }
                />
              </span>

              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Last trigger
                </Text>

                <ActionTriggerType
                  triggerType={
                    actionWithRecentRuns.runs?.[0]?.triggered_by_type
                  }
                  componentName={
                    actionWithRecentRuns.runs?.[0]?.run_env_vars?.COMPONENT_NAME
                  }
                  componentPath={`/${orgId}/installs/${installId}/components/${actionWithRecentRuns.runs?.[0]?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>
            </>
          ) : null}
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {actionWithRecentRuns?.action_workflow?.configs?.[0]?.triggers?.find(
            (t) => t.type === 'manual'
          ) ? (
            <ActionTriggerButton
              actionWorkflow={actionWithRecentRuns?.action_workflow}
              installId={installId}
              orgId={orgId}
              workflowConfigId={
                actionWithRecentRuns.action_workflow?.configs?.[0]?.id
              }
            />
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Recent executions">
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading
                  loadingText="Loading action run history..."
                  variant="stack"
                />
              }
            >
              <LoadRunHistory
                actionWorkflowId={actionWorkflowId}
                installId={installId}
                orgId={orgId}
                offset={sp['offset'] || '0'}
              />
            </Suspense>
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Latest configured steps">
            <div className="flex flex-col gap-2">
              {actionWithRecentRuns?.action_workflow?.configs?.[0]?.steps
                ?.sort((a, b) => b?.idx - a?.idx)
                ?.reverse()
                ?.map((s) => {
                  return (
                    <Expand
                      isOpen
                      id={s.id}
                      key={s.id}
                      parentClass="border rounded"
                      headerClass="px-3 py-2"
                      heading={<Text variant="med-12">{s.name}</Text>}
                      expandContent={
                        <div className="flex flex-col gap-4 p-3 border-t">
                          {s?.connected_github_vcs_config ||
                          s?.public_git_vcs_config ? (
                            <Config>
                              <ConfigurationVCS vcs={s} />
                            </Config>
                          ) : null}

                          {s.inline_contents?.length > 0 ? (
                            <div className="flex flex-col gap-2">
                              <Text variant="med-12">Inline contents</Text>
                              <CodeViewer initCodeSource={s.inline_contents} />
                            </div>
                          ) : null}

                          {s?.command?.length > 0 ? (
                            <div className="flex flex-col gap-2">
                              <Text variant="med-12">Command</Text>
                              <CodeViewer initCodeSource={s?.command} />
                            </div>
                          ) : null}

                          {s?.env_vars ? (
                            <ConfigurationVariables variables={s.env_vars} />
                          ) : null}
                        </div>
                      }
                    />
                  )
                })}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}

const LoadRunHistory = async ({
  actionWorkflowId,
  installId,
  orgId,
  limit = '6',
  offset,
}: {
  actionWorkflowId: string
  installId: string
  orgId: string
  limit?: string
  offset?: string
}) => {
  const params = new URLSearchParams({ offset, limit }).toString()
  const { data: actionWithRecentRuns, headers } =
    await nueQueryData<TInstallActionWorkflow>({
      orgId,
      path: `installs/${installId}/action-workflows/${actionWorkflowId}/recent-runs${params ? '?' + params : params}`,
      headers: {
        'x-nuon-pagination-enabled': true,
      },
    })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return actionWithRecentRuns ? (
    <div className="flex flex-col gap-4 w-full">
      <InstallWorkflowRunHistory
        orgId={orgId}
        installId={installId}
        actionsWithRecentRuns={actionWithRecentRuns}
        shouldPoll
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <Text>Unable to load action run history.</Text>
  )
}
