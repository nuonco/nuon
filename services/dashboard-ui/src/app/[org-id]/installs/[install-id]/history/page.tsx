import { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Loading,
  InstallHistory,
  InstallPageSubNav,
  InstallStatuses,
  InstallReprovisionButton,
  Section,
} from '@/components'
import { getInstall, getInstallEvents, getOrg } from '@/lib'
import { USER_REPROVISION } from '@/utils'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
          text: install.name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section
          heading="History"
          className="overflow-auto"
          actions={
            USER_REPROVISION ? (
              <InstallReprovisionButton installId={installId} orgId={orgId} />
            ) : null
          }
        >
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            <LoadInstallHistory installId={installId} orgId={orgId} />
          </Suspense>
        </Section>
      </div>
    </DashboardContent>
  )
})

const LoadInstallHistory: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const events = await getInstallEvents({ installId, orgId })
  return (
    <InstallHistory
      initEvents={events}
      installId={installId}
      orgId={orgId}
      shouldPoll
    />
  )
}
