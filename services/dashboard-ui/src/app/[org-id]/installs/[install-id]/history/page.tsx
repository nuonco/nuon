import { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Loading,
  InstallDeployComponentButton,
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
      statues={
        <div className="flex items-end gap-8">
          <InstallStatuses initInstall={install} shouldPoll />
          {USER_REPROVISION ? (
            <div className="flex items-center gap-3">
              <InstallReprovisionButton installId={installId} orgId={orgId} />
              {install?.install_components?.length ? (
                <InstallDeployComponentButton
                  installId={installId}
                  orgId={orgId}
                />
              ) : null}
            </div>
          ) : null}
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="History" className="overflow-auto">
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
