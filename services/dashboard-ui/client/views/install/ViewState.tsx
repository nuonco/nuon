import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallState } from '@/lib'
import { createFileDownload } from '@/utils/file-download'
import { CustomerManagedSnapshotState } from '@/components/customer-managed-support/SnapshotState'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export const ViewState = () => {
  const { install } = useInstall()

  return isCustomerManagedInstall(install) ? (
    <CustomerManagedSnapshotState />
  ) : (
    <ConnectedViewState />
  )
}

const ConnectedViewState = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [isCopied, setIsCopied] = useState(false)

  const {
    data: state,
    error,
    isLoading,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-state', org?.id, install?.id],
    queryFn: () => getInstallState({ orgId: org.id, installId: install.id }),
    enabled: !!org?.id && !!install?.id,
  })

  useEffect(() => {
    if (!isCopied) return
    const timeout = setTimeout(() => setIsCopied(false), 2000)
    return () => clearTimeout(timeout)
  }, [isCopied])

  return (
    <PageSection>
      <PageTitle segments={['State', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/state`,
            text: 'State',
          },
        ]}
      />
      <SectionHeader
        title="Install state"
        description="Raw state data for this install."
        actions={
          state ? (
            <>
              <Button
                variant="secondary"
                aria-label={isCopied ? 'Copied' : 'Copy state as JSON'}
                title={isCopied ? 'Copied' : 'Copy state as JSON'}
                onClick={() => {
                  navigator.clipboard.writeText(JSON.stringify(state, null, 2))
                  setIsCopied(true)
                }}
              >
                <Icon variant={isCopied ? 'CheckIcon' : 'CopyIcon'} size="16" />
              </Button>
              <Button
                variant="secondary"
                onClick={() =>
                  createFileDownload(
                    JSON.stringify(state, null, 2),
                    `${install?.name || 'install'}-state.json`,
                    'application/json'
                  )
                }
              >
                <Icon variant="DownloadSimpleIcon" size="16" />
                Download
              </Button>
            </>
          ) : null
        }
      />

      {error ? (
        <Banner theme="error">
          {(error as any)?.error || 'Unable to load install state.'}
        </Banner>
      ) : null}

      {isLoading ? (
        <Skeleton height="458px" width="100%" />
      ) : state ? (
        <JSONViewer
          className="min-h-[458px] max-h-[80vh] bg-code"
          data={state}
          expanded={1}
        />
      ) : (
        <div className="flex items-center justify-center p-8">
          <Text variant="body" theme="neutral">
            No state data available
          </Text>
        </div>
      )}
    </PageSection>
  )
}
