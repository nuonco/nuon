import type { ReactNode } from 'react'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCloudConnection } from '@/types'
import { ConnectionMetadata } from '../ConnectionDetail/ConnectionDetail'
import { ConnectionStatus } from '../ConnectionStatus'
import { isImagePreset } from '../repository-scope'

export const ConnectionWizard = ({
  connection,
  created = false,
  loading = false,
  children,
}: {
  connection?: TCloudConnection
  created?: boolean
  loading?: boolean
  children: ReactNode
}) =>
  connection || created || loading ? (
    <DetailPage
      header={
        <DetailHeader
          backLink={false}
          loading={loading}
          title={connection?.name || 'Cloud connection'}
          description={
            connection && isImagePreset(connection.preset)
              ? 'Configure cloud trust, then optionally generate repository access commands.'
              : 'Configure and verify the principal.'
          }
          status={<ConnectionStatus connection={connection} />}
          metadata={<ConnectionMetadata connection={connection} />}
        />
      }
    >
      {children}
    </DetailPage>
  ) : (
    <PageSection>
      <SectionHeader
        title="Create cloud connection"
        description="Connect a principal without storing credentials in Nuon."
      />
      {children}
    </PageSection>
  )
