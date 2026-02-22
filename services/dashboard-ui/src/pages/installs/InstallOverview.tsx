import { useParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useQuery } from '@/hooks/use-query'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { Markdown } from '@/components/common/Showdown'
import { Notice, Text, Loading, Section, InstallInputs, InstallInputsModal, SectionHeader } from '@/components'
import type { TInstallReadme, TInstallCurrentInputs } from '@/types'

const Readme = ({ installId, orgId }: { installId: string; orgId: string }) => {
  const { data: installReadme, error, isLoading } = useQuery<TInstallReadme>({
    path: `/api/ctl-api/v1/installs/${installId}/readme`,
  })

  if (isLoading) {
    return <Loading variant="stack" loadingText="Loading install README..." />
  }

  return installReadme && !error ? (
    <div className="flex flex-col gap-3">
      {installReadme?.warnings?.length
        ? installReadme?.warnings?.map((warn, i) => (
            <Notice key={i.toString()} variant="warn">
              {warn}
            </Notice>
          ))
        : null}
      <Markdown content={installReadme?.readme} />
    </div>
  ) : (
    <Text variant="reg-12">No install README found</Text>
  )
}

const CurrentInputs = ({ installId, orgId }: { installId: string; orgId: string }) => {
  const { data: currentInputs, isLoading } = useQuery<TInstallCurrentInputs>({
    path: `/api/ctl-api/v1/installs/${installId}/current-inputs`,
  })

  if (isLoading) {
    return <Loading variant="stack" loadingText="Loading install inputs..." />
  }

  return (
    <>
      <SectionHeader
        actions={
          currentInputs?.redacted_values ? (
            <InstallInputsModal currentInputs={currentInputs} />
          ) : undefined
        }
        className="mb-4"
        heading="Current inputs"
      />
      {currentInputs?.redacted_values ? (
        <InstallInputs currentInputs={currentInputs} />
      ) : (
        <Text>No inputs configured.</Text>
      )}
    </>
  )
}

export default function InstallOverview() {
  const { orgId, installId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection className="!pt-0" isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name || '',
          },
        ]}
      />
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <Section
          heading="README"
          className="md:col-span-8 !p-0"
          headingClassName="px-6 pt-6"
          childrenClassName="overflow-auto px-6 pb-6"
        >
          <Readme installId={installId || ''} orgId={orgId || ''} />
        </Section>

        <div className="divide-y flex flex-col col-span-4">
          <Section className="flex-initial">
            <CurrentInputs installId={installId || ''} orgId={orgId || ''} />
          </Section>
        </div>
      </div>
    </PageSection>
  )
}
