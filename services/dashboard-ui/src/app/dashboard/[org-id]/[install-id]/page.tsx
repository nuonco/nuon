import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { getInstall, getInstallEvents } from '@/lib'
import { ClientRefechInstallPage } from './install'

export default withPageAuthRequired(
  async function InstallDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string

    const [install, events] = await Promise.all([
      getInstall({ installId, orgId }),
      getInstallEvents({ installId, orgId }),
    ])

    return <ClientRefechInstallPage {...{ install, events }} />
  },
  { returnTo: '/dashboard' }
)
