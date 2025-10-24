import type { Metadata } from 'next'
import Image from 'next/image'
import { DashboardContent, Section, Text } from '@/components'
import { BreakGlassForm } from '@/components/old/Installs'
import { getInstallById } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install} = await getInstallById({ installId, orgId })

  return {
    title: `Generate break glass stack | ${install.name} | Nuon`,
  }
}

export default async function InstallBreakGlass({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const {data: install} = await getInstallById({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        {
          href: `/${orgId}/installs`,
          text: 'Installs',
        },
        { href: `/${orgId}/installs/${installId}`, text: install.name },
        {
          href: `/${orgId}/installs/${installId}/break-glass`,
          text: 'Access permissions',
        },
      ]}
    >
      <>
        <header className="px-6 py-8 flex flex-col border-b">
          <hgroup className="flex flex-col gap-2">
            <Text level={1} role="heading" variant="semi-18">
              Break glass permissions
            </Text>
            <Text
              variant="reg-12"
              className="text-cool-grey-600 dark:text-white/70"
            >
              Develop robust CloudFormation templates to streamline AWS
              infrastructure deployment for clients.
            </Text>
          </hgroup>
        </header>

        <Section
          heading={
            <div className="flex flex-col gap-2">
              <Text variant="semi-18">
                <Image
                  className=""
                  src={`/aws-cloudformation.svg`}
                  alt=""
                  height={24}
                  width={24}
                />
                CloudFormation
              </Text>
              <Text
                variant="reg-14"
                className="text-cool-grey-600 dark:text-white/70"
              >
                Note: Review access permissions thorougly before implementing
                modifications.
              </Text>
            </div>
          }
        >
          <BreakGlassForm install={install} />
        </Section>
      </>
    </DashboardContent>
  )
}
