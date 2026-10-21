import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router'
import { CreateInstallWizard } from '@/components/installs/CreateInstallWizard'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useInstallUIWizards } from '@/hooks/use-install-ui-wizards'
import { useOrg } from '@/hooks/use-org'
import { BranchProvider } from '@/providers/branch-provider'
import { createInstallWizardPath } from '@/components/installs/CreateInstallWizard/wizard-path'

const Redirect = ({ to }: { to: string }) => {
  const navigate = useNavigate()
  useEffect(() => {
    navigate(to, { replace: true })
  }, [navigate, to])
  return null
}

const CreateInstallWizardView = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const branchId = params.branchId as string
  const basePath = `/${org.id}/apps/${app.id}/branches/${branchId}`
  const installsPath = `${basePath}/installs`

  return (
    <>
      <PageTitle segments={['Create install', app.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/apps`, text: 'Apps' },
          { path: `/${org.id}/apps/${app.id}`, text: app.name },
          { path: basePath, text: branch.name },
          {
            path: createInstallWizardPath(org.id, app.id, branchId),
            text: 'Create install',
          },
        ]}
      />
      <CreateInstallWizard installsPath={installsPath} />
    </>
  )
}

export const CreateInstallWizardPage = () => {
  const enabled = useInstallUIWizards()
  const { org } = useOrg()
  const { app } = useApp()
  const params = useParams()
  const branchId = params.branchId as string
  const installsPath = `/${org.id}/apps/${app.id}/branches/${branchId}/installs`

  if (!enabled) {
    return (
      <>
        <PageTitle segments={['Create install', app.name]} />
        <Redirect to={installsPath} />
      </>
    )
  }

  return (
    <BranchProvider branchId={branchId}>
      <CreateInstallWizardView />
    </BranchProvider>
  )
}
