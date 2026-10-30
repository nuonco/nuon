import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCloudConnection } from '@/types'
import { ConnectionMetadata } from '../ConnectionDetail/ConnectionDetail'
import { ConnectionStatus } from '../ConnectionStatus'

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
          title={
            <Text
              family={connection ? 'mono' : 'sans'}
              weight="strong"
              className="break-all"
            >
              {connection?.name || 'Cloud connection'}
            </Text>
          }
          description={
            connection && ['ecr', 'gar', 'acr'].includes(connection.preset)
              ? 'Configure cloud trust, then optionally generate repository access commands.'
              : 'Configure and verify the cloud identity.'
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
        description="Connect a cloud identity without storing credentials in Nuon."
      />
      {children}
    </PageSection>
  )
