import { useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { usePolling } from '@/hooks/use-polling'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { InstallComponentHeader } from '@/components/install-components/InstallComponentHeader'
import { BackToTop } from '@/components/common/BackToTop'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import type { TInstallComponent } from '@/types'

export default function InstallComponentDetail() {
  const { org } = useOrg()
  const { install } = useInstall()
  const { componentId, orgId, installId } = useParams()

  const { data: installComponent, isLoading } = usePolling<TInstallComponent>({
    path: `/api/ctl-api/v1/installs/${installId}/components/${componentId}`,
    pollInterval: 20000,
    shouldPoll: true,
  })

  if (isLoading) {
    return (
      <PageSection isScrollable>
        <Loading variant="stack" loadingText="Loading component details..." />
      </PageSection>
    )
  }

  if (!installComponent) {
    return (
      <PageSection isScrollable>
        <Text theme="neutral">Component not found.</Text>
      </PageSection>
    )
  }

  const latestDeploy = installComponent?.install_deploys?.[0]

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${orgId}`, text: org?.name || '' },
          { path: `/${orgId}/installs`, text: 'Installs' },
          { path: `/${orgId}/installs/${installId}`, text: install?.name || '' },
          { path: `/${orgId}/installs/${installId}/components`, text: 'Components' },
          {
            path: `/${orgId}/installs/${installId}/components/${componentId}`,
            text: installComponent?.component?.name || 'Component Detail',
          },
        ]}
      />
      {latestDeploy ? (
        <InstallComponentHeader
          initDeploy={latestDeploy}
          installComponent={installComponent}
          shouldPoll={true}
        />
      ) : (
        <Text theme="neutral">No deploys found for this component.</Text>
      )}
      <BackToTop />
    </PageSection>
  )
}
