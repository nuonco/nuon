import { useOutletContext, useParams } from 'react-router'
import { Card } from '@/components/common/Card'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentOverviewTab = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    installComponent,
    isLoading,
    latestDeploy,
  } = useOutletContext<TInstallComponentOutletContext>()

  const component = installComponent?.component

  return (
    <>
      <PageTitle segments={[component?.name ?? 'Component', install?.name]} />

      <div className="flex flex-col gap-4">
        <SectionHeader title="Latest deploy" />
        <LatestDeployCard
          deploy={latestDeploy}
          isLoading={isLoading}
          href={
            latestDeploy?.id
              ? `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${latestDeploy.id}`
              : undefined
          }
        />
      </div>

      {component ? (
        <div className="flex flex-col gap-4">
          <SectionHeader title="Health" />
          <Card>
            <HealthTimeline installComponentId={componentId} shouldPoll />
          </Card>
        </div>
      ) : null}
    </>
  )
}
