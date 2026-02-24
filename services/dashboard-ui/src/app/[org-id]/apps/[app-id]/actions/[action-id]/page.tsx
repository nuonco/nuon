// NOTE(nnnat): needs refactored to stratus

import cronstrue from 'cronstrue'
import type { Metadata } from 'next'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { Code } from '@/components/common/Code'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAction, getOrg } from '@/lib'

// NOTE: old layout stuff
import {
  ActionTriggerType,
  CodeViewer,
  Config,
  ConfigurationVariables,
  ConfigurationVCS,
  Expand,
  Section,
  Text as OldText,
} from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['action-id']: actionId,
  } = await params
  const [{ data: app }, { data: action }] = await Promise.all([
    getApp({ appId, orgId }),
    getAction({ actionId, appId, orgId }),
  ])

  return {
    title: `${action.name} | Actions | ${app.name} | Nuon`,
  }
}

export default async function AppActionPage({ params }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['action-id']: actionId,
  } = await params
  const [{ data: app }, { data: action }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getAction({ actionId, appId, orgId }),
    getOrg({ orgId }),
  ])

  const containerId = 'app-action-page'
  return (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/actions`,
            text: 'Actions',
          },
          {
            path: `/${orgId}/apps/${appId}/actions/${actionId}`,
            text: action?.name,
          },
        ]}
      />
      {/* old page layout */}
      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {action?.name}
          </Text>
          <ID>{actionId}</ID>
        </HeadingGroup>
      </div>

      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Steps">
          <div className="flex flex-col gap-4">
            {action.configs[0].steps
              ?.sort((a, b) => b?.idx - a?.idx)
              ?.reverse()
              ?.map((s, i) => {
                return (
                  <Expand
                    isOpen
                    id={s.id}
                    key={s.id}
                    parentClass="border rounded"
                    headerClass="px-3 py-2"
                    heading={
                      <OldText variant="med-12">
                        {i + 1}. {s.name}
                      </OldText>
                    }
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
                            <OldText variant="med-12">Inline contents</OldText>
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

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Triggers">
            <div className="flex flex-col divide-y">
              {action.configs[0].triggers.map((t) => (
                <div className="flex gap-2 py-2" key={t.id}>
                  <ActionTriggerType
                    triggerType={t.type}
                    componentName={t?.component?.name}
                    componentPath={`/${orgId}/apps/${appId}/components/${t?.component_id}`}
                  />
                  {t.type === 'cron' ? (
                    <OldText variant="reg-12">
                      Will run{' '}
                      {cronstrue
                        .toString(t.cron_schedule, { verbose: true })
                        .toLowerCase()}
                      .
                    </OldText>
                  ) : null}
                </div>
              ))}
            </div>
          </Section>
          {action.configs?.[0]?.break_glass_role_arn!! ? (
            <Section
              className="flex-initial"
              childrenClassName="flex flex-col gap-4"
              heading="Break glass role"
            >
              <Text>
                Role{' '}
                <Code variant="inline">
                  {action?.configs?.[0]?.break_glass_role_arn}
                </Code>{' '}
                must be enabled in install stack before running this action.
              </Text>
            </Section>
          ) : null}
          {action.configs?.[0]?.role && (
            <Section
              className="flex-initial"
              childrenClassName="flex flex-col gap-4"
              heading="Execution Role"
            >
              <Text variant="body" level={5}>
                Configured Role
              </Text>
              <Text variant="subtext">
                IAM role used when executing this action.
              </Text>
              <Code variant="inline">{action.configs[0].role}</Code>
            </Section>
          )}
        </div>
      </div>
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
