import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { Code } from '@/components/common/Code'
import { Cron } from '@/components/common/Cron'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAction } from '@/lib'
import type { TActionConfigTriggerType } from "@/types"

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
      <PageTitle title={`${action?.name ?? 'Action'} | ${app?.name}`} />
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

      <div className="p-6 border-b flex items-start justify-between gap-8">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {action?.name}
          </Text>
          {actionId ? <ID>{actionId}</ID> : null}
        </HeadingGroup>

        {config && (config.triggers?.length || config.break_glass_role_arn || config.role) ? (
          <div className="flex flex-row gap-6 items-start">
            {config.triggers?.length ? (
              <LabeledValue label="Triggers">
                <div className="flex flex-col gap-2">
                  {config.triggers.map((trigger) => (
                    <div key={trigger.id} className="flex items-center gap-2 flex-wrap">
                      <ActionTriggerType
                        triggerType={trigger.type as TActionConfigTriggerType}
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
                    </div>
                  ))}
                </div>
              </LabeledValue>
            ) : null}

            {config.break_glass_role_arn ? (
              <LabeledValue label="Break glass role">
                <Code variant="inline">{config.break_glass_role_arn}</Code>
                <Text variant="subtext" theme="neutral">
                  Must be enabled in the install stack before running this action.
                </Text>
              </LabeledValue>
            ) : null}

            {config.role ? (
              <LabeledValue label="Execution role">
                <Code variant="inline">{config.role}</Code>
              </LabeledValue>
            ) : null}
          </div>
        ) : null}
      </div>

      <PageSection>
        <Text variant="base" weight="strong">
          Steps
        </Text>
        {steps?.length ? (
          <div className="grid grid-cols-1 gap-4">
            {steps.map((step, i) => (
              <ActionStep key={step.id} step={step} index={i} />
            ))}
          </div>
        ) : (
          <Text theme="neutral">No steps configured.</Text>
        )}
      </PageSection>

      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
