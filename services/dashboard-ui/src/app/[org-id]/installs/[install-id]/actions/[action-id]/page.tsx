import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getInstallActionById, getInstallById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { ActionRuns } from './action-runs'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  ActionTriggerButton,
  ActionTriggerType,
  ClickToCopy,
  CodeViewer,
  Config,
  ConfigurationVCS,
  ConfigurationVariables,
  ErrorFallback,
  Expand,
  Loading,
  DashboardContent,
  Section,
  StatusBadge,
  Text as OldText,
  ToolTip,
  Truncate,
} from '@/components'

type TInstallPageProps = TPageProps<'org-id' | 'install-id' | 'action-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
  } = await params
  const [{ data: install }, { data: installAction }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getInstallActionById({ actionId, installId, orgId }),
  ])

  return {
    title: `${install.name} | ${installAction.action_workflow?.name}`,
  }
}

export default async function InstallActionPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
  } = await params
  const sp = await searchParams
  const [{ data: install }, { data: installAction }, { data: org }] =
    await Promise.all([
      getInstallById({ installId, orgId }),
      getInstallActionById({ actionId, installId, orgId }),
      getOrgById({ orgId }),
    ])

  const containerId = 'install-action-page'

  return org?.features?.['stratus-layout'] ? (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      {/* old page layout */}

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {installAction.action_workflow?.name}
          </Text>
          <ID>{actionId}</ID>
        </HeadingGroup>

        <div className="flex gap-6 items-start justify-start">
          {installAction?.runs?.length ? (
            <>
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Recent status
                </OldText>
                <StatusBadge
                  description={
                    installAction.runs?.[0]?.status_v2
                      ?.status_human_description ||
                    installAction?.runs?.[0]?.status_description
                  }
                  status={
                    installAction.runs?.[0]?.status_v2?.status ||
                    installAction?.runs?.[0]?.status ||
                    'noop'
                  }
                />
              </span>

              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Last trigger
                </OldText>

                <ActionTriggerType
                  triggerType={installAction.runs?.[0]?.triggered_by_type}
                  componentName={
                    installAction.runs?.[0]?.run_env_vars?.COMPONENT_NAME
                  }
                  componentPath={`/${orgId}/installs/${installId}/components/${installAction.runs?.[0]?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>
            </>
          ) : null}
          <span className="flex flex-col gap-2">
            <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </OldText>
            <OldText variant="med-12">{install.name}</OldText>
            <OldText variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate className="text-xs" variant="small">
                    {install.id}
                  </Truncate>
                </ClickToCopy>
              </ToolTip>
            </OldText>
          </span>
          {installAction?.action_workflow?.configs?.[0]?.triggers?.find(
            (t) => t.type === 'manual'
          ) ? (
            <ActionTriggerButton
              action={installAction?.action_workflow}
              actionConfigId={installAction.action_workflow?.configs?.[0]?.id}
            />
          ) : null}
        </div>
      </div>

      <div className="flex flex-col md:flex-row flex-auto md:divide-x">
        <Section heading="Recent executions">
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading
                  loadingText="Loading action run history..."
                  variant="stack"
                />
              }
            >
              <ActionRuns
                actionId={actionId}
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
              {installAction?.action_workflow?.configs?.[0]?.steps
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
                      heading={<OldText variant="med-12">{s.name}</OldText>}
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
                              <OldText variant="med-12">
                                Inline contents
                              </OldText>
                              <CodeViewer initCodeSource={s.inline_contents} />
                            </div>
                          ) : null}

                          {s?.command?.length > 0 ? (
                            <div className="flex flex-col gap-2">
                              <OldText variant="med-12">Command</OldText>
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
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
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
          href: `/${orgId}/installs/${install.id}/actions/${actionId}`,
          text: installAction?.action_workflow?.name,
        },
      ]}
      heading={installAction.action_workflow?.name}
      headingUnderline={installAction.action_workflow?.id}
      statues={
        <div className="flex gap-6 items-start justify-start">
          {installAction?.runs?.length ? (
            <>
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Recent status
                </OldText>
                <StatusBadge
                  description={
                    installAction.runs?.[0]?.status_v2
                      ?.status_human_description ||
                    installAction?.runs?.[0]?.status_description
                  }
                  status={
                    installAction.runs?.[0]?.status_v2?.status ||
                    installAction?.runs?.[0]?.status ||
                    'noop'
                  }
                />
              </span>

              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Last trigger
                </OldText>

                <ActionTriggerType
                  triggerType={installAction.runs?.[0]?.triggered_by_type}
                  componentName={
                    installAction.runs?.[0]?.run_env_vars?.COMPONENT_NAME
                  }
                  componentPath={`/${orgId}/installs/${installId}/components/${installAction.runs?.[0]?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>
            </>
          ) : null}
          <span className="flex flex-col gap-2">
            <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </OldText>
            <OldText variant="med-12">{install.name}</OldText>
            <OldText variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </OldText>
          </span>
          {installAction?.action_workflow?.configs?.[0]?.triggers?.find(
            (t) => t.type === 'manual'
          ) ? (
            <ActionTriggerButton
              action={installAction?.action_workflow}
              actionConfigId={installAction.action_workflow?.configs?.[0]?.id}
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
              <ActionRuns
                actionId={actionId}
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
              {installAction?.action_workflow?.configs?.[0]?.steps
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
                      heading={<OldText variant="med-12">{s.name}</OldText>}
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
                              <OldText variant="med-12">
                                Inline contents
                              </OldText>
                              <CodeViewer initCodeSource={s.inline_contents} />
                            </div>
                          ) : null}

                          {s?.command?.length > 0 ? (
                            <div className="flex flex-col gap-2">
                              <OldText variant="med-12">Command</OldText>
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
