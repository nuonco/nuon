import { useOutletContext, useParams } from 'react-router'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentDeploysTab = () => {
  const { componentId } = useParams()
  const { install } = useInstall()
  const { installComponent } =
    useOutletContext<TInstallComponentOutletContext>()
  const component = installComponent?.component

  return (
    <>
      <PageTitle segments={['Deploy history', install?.name]} />
      {component ? (
        <DeployTimeline
          componentId={componentId!}
          componentName={component.name}
          shouldPoll
        />
      ) : null}
    </>
  )
}
