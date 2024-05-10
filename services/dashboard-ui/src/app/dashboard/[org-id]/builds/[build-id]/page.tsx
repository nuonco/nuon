import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { getInstall, getInstallEvents } from '@/lib'
import { RTInstallPage } from "./rt-install"

export default withPageAuthRequired(
  async function Install2Dashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['build-id'] as string

    const [install, events] = await Promise.all([
      getInstall({ installId, orgId }),
      getInstallEvents({ installId, orgId }),
    ])

    return (
      <RTInstallPage  {...{ install, events }}/>
    )
  },
  { returnTo: '/dashboard' }
)
