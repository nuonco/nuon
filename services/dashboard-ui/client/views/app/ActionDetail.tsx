import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { Code } from '@/components/common/Code'
import { Cron } from '@/components/common/Cron'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAction } from '@/lib'

const CONTAINER_ID = 'action-detail-page'

export const ActionDetail = () => {
  const { actionId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: action } = useQuery({
    queryKey: ['action', org?.id, app?.id, actionId],
    queryFn: () =>
      getAction({ orgId: org.id, appId: app.id, actionId: actionId! }),
    enabled: !!org?.id && !!app?.id && !!actionId,
  })

  const config = action?.configs?.[0]
  const steps = config?.steps
    ?.slice()
    .sort((a, b) => (a?.idx ?? 0) - (b?.idx ?? 0))

  return (
    <PageSection id={CONTAINER_ID} isScrollable className="!p-0 !gap-0">
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/actions`, text: 'Actions' },
          {
            path: `/${org?.id}/apps/${app?.id}/actions/${actionId}`,
            text: action?.name,
          },
        ]}
      />

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {action?.name}
          </Text>
          {actionId ? <ID>{actionId}</ID> : null}
        </HeadingGroup>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <PageSection className="md:col-span-8">
          <Text variant="base" weight="strong">
            Steps
          </Text>
          {steps?.length ? (
            <div className="flex flex-col gap-4">
              {steps.map((step, i) => (
                <ActionStep key={step.id} step={step} index={i} />
              ))}
            </div>
          ) : (
            <Text theme="neutral">No steps configured.</Text>
          )}
        </PageSection>

        <PageSection className="md:col-span-4">
          <div className="flex flex-col divide-y gap-0">
            {config?.triggers?.length ? (
              <div className="flex flex-col gap-3 pb-6">
                <Text variant="base" weight="strong">
                  Triggers
                </Text>
                <div className="flex flex-col gap-3">
                  {config.triggers.map((trigger) => (
                    <div key={trigger.id} className="flex flex-col gap-1">
                      <span className="flex items-center gap-2">
                        <ActionTriggerType
                          triggerType={trigger.type}
                          componentName={trigger?.component?.name}
                          componentPath={`/${org?.id}/apps/${app?.id}/components/${trigger?.component_id}`}
                        />
                        {trigger.type === 'cron' && trigger.cron_schedule ? (
                          <Cron
                            cron={trigger.cron_schedule}
                            variant="subtext"
                            theme="neutral"
                            showTooltip
                          />
                        ) : null}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}

            {config?.break_glass_role_arn ? (
              <div className="flex flex-col gap-3 py-6">
                <Text variant="base" weight="strong">
                  Break glass role
                </Text>
                <Text variant="body">
                  Role{' '}
                  <Code variant="inline">{config.break_glass_role_arn}</Code>{' '}
                  must be enabled in install stack before running this action.
                </Text>
              </div>
            ) : null}

            {config?.role ? (
              <div className="flex flex-col gap-3 py-6">
                <Text variant="base" weight="strong">
                  Execution role
                </Text>
                <Text variant="subtext" theme="neutral">
                  IAM role used when executing this action.
                </Text>
                <Code variant="inline">{config.role}</Code>
              </div>
            ) : null}
          </div>
        </PageSection>
      </div>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
