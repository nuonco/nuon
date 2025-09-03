import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface ITeardownInstallComponents extends IGetInstall {}

export async function teardownInstallComponents({
  installId,
  orgId,
}: ITeardownInstallComponents) {
  return mutateData<string>({
    errorMessage: 'Unable to teardown install components.',
    orgId,
    path: `installs/${installId}/components/teardown-all`,
  })
}