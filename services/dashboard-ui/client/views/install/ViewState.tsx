import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallState } from '@/lib'

export const ViewState = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: state, error, isLoading } = useQuery({
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <PageSection>
      <PageTitle title={`State | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/state`,
            text: 'State',
          },
        ]}
      />
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install state
          </Text>
          <Text variant="subtext" theme="neutral">
            Raw state data for this install.
          </Text>
        </HeadingGroup>
        {state && (
          <ClickToCopyButton
            textToCopy={JSON.stringify(state, null, 2)}
            className="w-fit"
          />
        )}
      </div>

      {error ? (
        <Banner theme="error">
          {(error as any)?.error || 'Unable to load install state.'}
        </Banner>
      ) : null}

      {isLoading ? (
        <Skeleton height="458px" width="100%" />
      ) : state ? (
        <JSONViewer
          className="min-h-[458px] max-h-[80vh] bg-code"
          data={state}
          expanded={1}
        />
      ) : (
        <div className="flex items-center justify-center p-8">
          <Text variant="body" theme="neutral">
            No state data available
          </Text>
        </div>
      )}
    </PageSection>
  )
}
