import { type ReactNode } from 'react'
import { Navigate } from 'react-router'
import { Spinner } from '../components/atoms/Spinner'
import { Text } from '../components/atoms/Text'
import { installSetupHref } from '../utils/hrefs'
import {
  installSetupDescriptor,
  installSetupStateFromInstall,
} from '../utils/install-setup'
import { isWizardComplete } from '../utils/wizard'
import { useInstall } from '../providers/install-provider'
import { useOrg } from '../providers/org-provider'

export const InstallResolver = ({ children }: { children: ReactNode }) => {
  const { orgId } = useOrg()
  const { install, installId, loading, error } = useInstall()

  if (!orgId || !installId || loading) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <Spinner size={20} label="Loading install" />
      </div>
    )
  }

  if (error) {
    return (
      <Text variant="caption" color="secondary">
        Install failed to load
      </Text>
    )
  }

  if (
    !isWizardComplete(
      installSetupDescriptor,
      installSetupStateFromInstall(install)
    )
  ) {
    return <Navigate replace to={installSetupHref(orgId, installId)} />
  }

  return children
}
