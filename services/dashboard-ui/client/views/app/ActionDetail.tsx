import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { Duration } from '@/components/common/Duration'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { ActionStep } from '@/components/actions/ActionStep'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAction } from '@/lib'
import type { TActionConfigTriggerType } from '@/types'
import { sortByIdx } from '@/utils/action-utils'

export const ActionDetail = () => {
  const { actionId, branchId } = useParams()
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`

  const { data: action, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['action', org?.id, app?.id, actionId],
    queryFn: () =>
      getAction({ orgId: org.id, appId: app.id, actionId: actionId! }),
    enabled: !!org?.id && !!app?.id && !!actionId,
  })

  const config = action?.configs?.[0]
  const steps = config?.steps ? sortByIdx(config.steps) : undefined
  const labelKeys = Object.keys(action?.labels ?? {}).sort()

  return (
    <>
      <PageTitle segments={[action?.name ?? 'Action', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `${appBase}/actions`, text: 'Actions' },
          {
            path: `${appBase}/actions/${actionId}`,
            text: action?.name,
          },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            title={action?.name}
            loading={isLoading}
            loadingWidth={20}
            id={actionId}
            identity={
              labelKeys.length ? (
                <span className="flex flex-wrap gap-1">
                  {labelKeys.map((k) => (
                    <LabelBadge
                      key={k}
                      labelKey={k}
                      labelValue={action?.labels?.[k]}
                      size="sm"
                      customColor={labelColors?.[k]}
                    />
                  ))}
                </span>
              ) : null
            }
            metadata={
              isLoading ? (
                <>
                  <LabeledValue label="Timeout" loading />
                  <LabeledValue label="Kube config" loading />
                  <LabeledValue label="Triggers" loading />
                </>
              ) : config ? (
                <>
                  {config?.timeout ? (
                    <LabeledValue label="Timeout">
                      <Duration
                        nanoseconds={config?.timeout}
                        variant="subtext"
                      />
                    </LabeledValue>
                  ) : null}

                  <LabeledValue label="Kube config">
                    <Badge
                      theme={config?.enable_kube_config ? 'info' : 'warn'}
                      variant="code"
                      size="sm"
                    >
                      {config?.enable_kube_config ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </LabeledValue>

                  {config?.image ? (
                    <LabeledValue label="Container image">
                      <Code variant="inline">{config.image}</Code>
                    </LabeledValue>
                  ) : null}

                  {config.triggers?.length ? (
                    <LabeledValue label="Triggers">
                      <div className="flex flex-col gap-2">
                        {config.triggers.map((trigger) => (
                          <div
                            key={trigger.id}
                            className="flex items-center gap-2 flex-wrap"
                          >
                            <ActionTriggerType
                              size="sm"
                              triggerType={
                                trigger.type as TActionConfigTriggerType
                              }
                              componentName={trigger?.component?.name}
                              componentPath={`${appBase}/components/${trigger?.component_id}`}
                              cronSchedule={trigger?.cron_schedule}
                            />
                          </div>
                        ))}
                      </div>
                    </LabeledValue>
                  ) : null}

                  {config.break_glass_role_arn ? (
                    <LabeledValue label="Break glass role">
                      <Code variant="inline">
                        {config.break_glass_role_arn}
                      </Code>
                      <Text variant="subtext" theme="neutral">
                        Must be enabled in the install stack before running this
                        action.
                      </Text>
                    </LabeledValue>
                  ) : null}

                  {config.role ? (
                    <LabeledValue label="Execution role">
                      <Code variant="inline">{config.role}</Code>
                    </LabeledValue>
                  ) : null}
                </>
              ) : null
            }
          />
        }
      >
        <div className="flex flex-col gap-4">
          <SectionHeader title="Steps" />
          {steps?.length ? (
            <div className="grid grid-cols-1 gap-4">
              {steps.map((step, i) => (
                <ActionStep key={step.id} step={step} index={i} />
              ))}
            </div>
          ) : (
            <Text theme="neutral">No steps configured.</Text>
          )}
        </div>
      </DetailPage>
    </>
  )
}
