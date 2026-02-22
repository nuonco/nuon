import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { InstallComponentsTable } from '@/components/install-components/InstallComponentsTable'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TInstallComponent, TAppConfig, TAPIResponse } from '@/types'

const LIMIT = 10

const InstallComponentsTableWrapper = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'
  const q = searchParams.get('q') || ''
  const types = searchParams.get('types') || ''

  const {
    data: componentsResponse,
    error: componentsError,
    isLoading: componentsLoading,
    headers: componentsHeaders,
  } = usePolling<TInstallComponent[]>({
    path: `/api/ctl-api/v1/installs/${installId}/components?limit=${LIMIT}&offset=${offset}${q ? `&q=${q}` : ''}${types ? `&types=${types}` : ''}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const {
    data: configResponse,
    error: configError,
    isLoading: configLoading,
  } = useQuery<TAppConfig>({
    path: `/api/ctl-api/v1/apps/${install?.app_id}/configs/${install?.app_config_id}?recurse=true`,
  })

  const pagination = {
    limit: Number(componentsHeaders?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: componentsHeaders?.['x-nuon-page-next'] === 'true',
    offset: Number(componentsHeaders?.['x-nuon-page-offset'] ?? '0'),
  }

  const componentDeps =
    componentsResponse?.map((ic) => ({
      id: ic?.id,
      component_id: ic?.component_id,
      dependencies: configResponse?.component_config_connections?.find(
        (c) => c?.component_id === ic?.component_id
      )?.component_dependency_ids,
    })) || []

  if (componentsError && !componentsResponse) {
    return (
      <div>
        <p>Could not load your components.</p>
        <p>{componentsError.message || 'Unknown error'}</p>
      </div>
    )
  }

  if (componentsLoading && !componentsResponse) {
    return <div>Loading components...</div>
  }

  return (
    <InstallComponentsTable
      components={componentsResponse || []}
      deps={componentDeps}
      pagination={pagination}
      shouldPoll
    />
  )
}

export default function InstallComponents() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  if (!installId || !orgId) {
    return null
  }

  return (
    <PageSection isScrollable>
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
            path: `/${orgId}/installs/${installId}/components`,
            text: 'Components',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install components
        </Text>
        <Text theme="neutral">
          View and manage all components for this install.
        </Text>
      </HeadingGroup>

      <InstallComponentsTableWrapper installId={installId} orgId={orgId} />

      <BackToTop />
    </PageSection>
  )
}