import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getRunbook } from '@/lib'

export const RunbookDetail = () => {
  const { runbookId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: runbook } = useQuery({
    queryKey: ['runbook', org?.id, app?.id, runbookId],
    queryFn: () => getRunbook({ orgId: org!.id, appId: app!.id, runbookId: runbookId! }),
    enabled: !!org?.id && !!app?.id && !!runbookId,
  })

  const steps = runbook?.steps ?? []

  return (
    <PageSection flush>
      <PageTitle title={`${runbook?.name ?? 'Runbook'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
          {
            path: `/${org?.id}/apps/${app?.id}/runbooks/${runbookId}`,
            text: runbook?.name,
          },
        ]}
      />

      <div className="p-6 border-b">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {runbook?.name}
          </Text>
          {runbookId ? <ID>{runbookId}</ID> : null}
          {runbook?.description ? (
            <Text variant="subtext" theme="neutral">
              {runbook.description}
            </Text>
          ) : null}
        </HeadingGroup>
      </div>

      {runbook?.readme ? (
        <PageSection>
          <Markdown content={runbook.readme} mode="app" />
        </PageSection>
      ) : null}

      <PageSection>
        <Text variant="base" weight="strong">
          Steps
        </Text>
        {steps.length ? (
          <div className="grid grid-cols-1 gap-4">
            {steps.map((step, i) => (
              <div
                key={step.id ?? i}
                className="border rounded-lg p-4 flex flex-col gap-1"
              >
                <Text variant="body" weight="strong">
                  {i + 1}. {step.name}
                </Text>
                {step.description ? (
                  <Text variant="subtext" theme="neutral">
                    {step.description}
                  </Text>
                ) : null}
                {step.type ? (
                  <Text variant="subtext" theme="neutral">
                    Type: {step.type}
                  </Text>
                ) : null}
              </div>
            ))}
          </div>
        ) : (
          <Text theme="neutral">No steps configured.</Text>
        )}
      </PageSection>
    </PageSection>
  )
}
