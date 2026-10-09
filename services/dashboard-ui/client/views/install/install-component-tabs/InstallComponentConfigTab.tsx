import { useOutletContext, useParams } from 'react-router'
import { InstallComponentConfigBody } from '@/components/install-components/InstallComponentConfig'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstallPage } from '@/hooks/use-install-path'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentConfigTab = () => {
  const { componentId } = useParams()
  const { install } = useInstallPage()
  const context = useOutletContext<TInstallComponentOutletContext>()

  return (
    <>
      <PageTitle segments={['Component config', install?.name]} />
      <InstallComponentConfigBody componentId={componentId} {...context} />
    </>
  )
}
