// NOTE(nnnnat): refactor to stratus

import type { Metadata } from 'next'
import { Suspense } from 'react'
import { InstallActionManualRunButton } from '@/components/actions/InstallActionManualRun'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Code } from '@/components/common/Code'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstallAction, getInstall, getInstallState, getOrg } from '@/lib'
import { getSession } from '@/lib/auth-server'
import type { TPageProps } from '@/types'
import { Runs, RunsError, RunsSkeleton } from './runs'

// NOTE: old layout stuff
import {
  ActionTriggerType,
  ClickToCopy,
  CodeViewer,
  Config,
  ConfigurationVCS,
  ConfigurationVariables,
  Expand,
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
    getInstall({ installId, orgId }),
    getInstallAction({ actionId, installId, orgId }),
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
  const session = await getSession()
  const [
    { data: install },
    { data: installAction },
    { data: installState },
    { data: org },
  ] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallAction({ actionId, installId, orgId }),
    getInstallState({ installId, orgId }),
    getOrg({ orgId }),
  ])

  const containerId = 'install-action-page'

  const installActionBreakGlassRole = installAction.action_workflow?.configs?.[0].break_glass_role_arn
  const breakGlassRoleArns = installState?.install_stack?.outputs?.break_glass_role_arns

  return (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/actions`,
            text: 'Actions',
          },
          {
            path: `/${orgId}/installs/${installId}/actions/${actionId}`,
            text: installAction?.action_workflow?.name,
          },
        ]}
      />
      {/* old page layout */}

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {installAction.action_workflow?.name}
          </Text>
          <span className="flex items-center gap-4">
            <ID>{actionId}</ID>{' '}
            {session?.user?.email?.endsWith('@nuon.co') ? (
              <ID>{installAction?.id}</ID>
            ) : null}
          </span>
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
            <InstallActionManualRunButton
              action={installAction?.action_workflow}
              actionConfigId={installAction.action_workflow?.configs?.[0]?.id}
            />
          ) : null}
        </div>
      </div>

      <div className="flex flex-col md:flex-row flex-auto md:divide-x">
        <Section heading="Recent executions">
          <ErrorBoundary fallback={<RunsError />}>
            <Suspense fallback={<RunsSkeleton />}>
              <Runs
                actionId={actionId}
                installId={installId}
                orgId={orgId}
                offset={sp['offset'] || '0'}
              />
            </Suspense>
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          {installActionBreakGlassRole ? (
            <Section
              className="flex-initial"
              childrenClassName="flex flex-col gap-4"
              heading={
                <div className="flex justify-between items-center gap-4 w-full">
                  <span>Break glass role</span>
                  <StatusBadge
                    description={
                      breakGlassRoleArns[installActionBreakGlassRole]
                        ? 'Role provisioned'
                        : 'Role not provisioned'
                    }
                    status={
                      breakGlassRoleArns[installActionBreakGlassRole]
                        ? 'provisioned'
                        : 'not provisioned'
                    }
                  />
                </div>
              }
            >
              <br></br>
              {breakGlassRoleArns[installActionBreakGlassRole]!! ? (
                <div className="flex flex-col gap-2">
                  <Text variant="body" weight="strong">
                    Role assumed while running this action 
                  </Text>
                  <Code variant='default'>
                    {
                        breakGlassRoleArns[installActionBreakGlassRole]
                    }
                  </Code>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  <Text variant="body" weight="strong" level={5}>
                    Access Permissions
                  </Text>
                  <Text>
                    Break Glass Role must be enabled in install stack before running this action.
                    <Code variant="default">
                      {
                        installAction?.action_workflow?.configs?.[0]
                          ?.break_glass_role_arn
                      }
                    </Code>
                  </Text>
                </div>
              )
            }
            </Section>
          ) : null}
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
  )
}
