import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { TeardownInstallWizard } from '@/components/installs/TeardownInstallWizard'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstallPage } from '@/hooks/use-install-path'
import { useInstallUIWizards } from '@/hooks/use-install-ui-wizards'

export const TeardownInstallWizardPage = () => {
  const enabled = useInstallUIWizards()
  const navigate = useNavigate()
  const { org, install, href, breadcrumbs } = useInstallPage()
  const installHref = href()

  useEffect(() => {
    if (!enabled) navigate(installHref, { replace: true })
  }, [enabled, installHref, navigate])

  if (!enabled) {
    return <PageTitle segments={['Teardown', install.name]} />
  }

  return (
    <>
      <PageTitle segments={['Teardown', install.name]} />
      <Breadcrumbs
        breadcrumbs={breadcrumbs(
          [
            { path: `/${org.id}`, text: org.name },
            { path: `/${org.id}/installs`, text: 'Installs' },
            { path: installHref, text: install.name },
            { path: href('/teardown'), text: 'Teardown' },
          ],
          [{ suffix: '/teardown', text: 'Teardown' }]
        )}
      />
      <TeardownInstallWizard />
    </>
  )
}
