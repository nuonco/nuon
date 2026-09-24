import { useOutletContext, useParams } from 'react-router'
import { Card } from '@/components/common/Card'
import { Cron } from '@/components/common/Cron'
import { LabeledValue } from '@/components/common/LabeledValue'
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
    config,
    installComponent,
    isDisabled,
    isLoading,
    isToggleable,
    latestDeploy,
  } = useOutletContext<TInstallComponentOutletContext>()

  const component = installComponent?.component

  return (
    <>
      <PageTitle segments={[component?.name ?? 'Component', install?.name]} />

      <div className="flex items-start justify-between gap-6">
        {config?.drift_schedule ? (
          <LabeledValue label="Drift schedule">
            <Cron cron={config.drift_schedule} variant="subtext" />
          </LabeledValue>
        ) : (
          <div />
        )}

        {isToggleable && component ? (
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
        ) : null}
      </div>

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
