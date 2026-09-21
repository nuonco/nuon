import { useOutletContext, useParams } from 'react-router'
import { Card } from '@/components/common/Card'
import { Toggle } from '@/components/common/form/Toggle'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { ToggleComponentModalContainer } from '@/components/install-components/management/ToggleComponent/ToggleComponentContainer'
import { HealthTimeline } from '@/components/install-health/HealthTimeline'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentOverviewTab = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addModal } = useSurfaces()
  const {
    installComponent,
    isDisabled,
    isLoading,
    isToggleable,
    latestDeploy,
  } = useOutletContext<TInstallComponentOutletContext>()

  const component = installComponent?.component
  const showHealth = !!org?.features?.['component-health']

  return (
    <>
      <PageTitle segments={[component?.name ?? 'Component', install?.name]} />

      {isToggleable && component ? (
        <div className="flex justify-end">
          <Toggle
            checked={!isDisabled}
            onChange={() => {
              addModal(
                <ToggleComponentModalContainer
                  component={component}
                  enabling={isDisabled}
                />
              )
            }}
            label={isDisabled ? 'Component disabled' : 'Component enabled'}
            description={
              isDisabled
                ? `${component.name} is disabled on this install. Toggle to deploy.`
                : `${component.name} can be disabled on this install.`
            }
          />
        </div>
      ) : null}

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

      {showHealth && component ? (
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
