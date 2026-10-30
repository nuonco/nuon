import { ConnectionDetail } from './ConnectionDetail'
import { InstallsTableComponent } from '@/components/installs/InstallsTable'
import { RepositoryAccessDemo } from '../__stories__/RepositoryAccessDemo'
import {
  PageStory,
  acrConnection,
  failedACRConnection,
  failedACRTrustConnection,
  verifiedACRConnection,
  connection,
  customConnection,
  ecrConnection,
  garConnection,
  failedGARConnection,
  failedGARTrustConnection,
  verifiedGARConnection,
  failedECRConnection,
  failedConnection,
  verifiedECRConnection,
  verifiedConnection,
  noop,
} from '../__stories__/fixtures'
export default { title: 'Features / Cloud connections / Detail' }
const props = {
  basePath: '/org-mock-001/settings/cloud-connections/cc-example',
  onVerify: noop,
  onDelete: noop,
}
export const Overview = () => (
  <PageStory>
    <ConnectionDetail {...props} connection={connection} tab="overview" />
  </PageStory>
)
export const Custom = () => (
  <PageStory>
    <ConnectionDetail {...props} connection={customConnection} tab="overview" />
  </PageStory>
)
export const ECROverview = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={ecrConnection}
      tab="overview"
      repositoryAccess={<RepositoryAccessDemo connection={ecrConnection} />}
    />
  </PageStory>
)
export const ECRVerification = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedECRConnection}
      tab="verification"
    />
  </PageStory>
)
export const ImageConfigurations = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={{
        ...verifiedECRConnection,
        used_by: { installs: 0, image_configs: 3 },
      }}
      tab="overview"
      repositoryAccess={
        <RepositoryAccessDemo connection={verifiedECRConnection} />
      }
    />
  </PageStory>
)
export const ECRTokenFailed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedECRConnection}
      tab="verification"
    />
  </PageStory>
)
export const Loading = () => (
  <PageStory>
    <ConnectionDetail {...props} tab="overview" />
  </PageStory>
)
export const InstallsEmpty = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={connection}
      tab="installs"
      installs={
        <InstallsTableComponent
          data={[]}
          isLoading={false}
          pagination={{ offset: 0, limit: 20 }}
          emptyTitle="No installs use this connection yet"
          emptyMessage="Select this connection when creating an install."
        />
      }
    />
  </PageStory>
)
export const Verification = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedConnection}
      tab="verification"
    />
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedConnection}
      tab="verification"
    />
  </PageStory>
)
export const Verifying = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedConnection}
      tab="verification"
      isVerifying
    />
  </PageStory>
)
export const GAROverview = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={garConnection}
      tab="overview"
      repositoryAccess={<RepositoryAccessDemo connection={garConnection} />}
    />
  </PageStory>
)
export const GARVerification = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedGARConnection}
      tab="verification"
    />
  </PageStory>
)
export const GARVerifying = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={garConnection}
      tab="verification"
      isVerifying
    />
  </PageStory>
)
export const GARImpersonationFailed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedGARConnection}
      tab="verification"
    />
  </PageStory>
)
export const GARTrustFailed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedGARTrustConnection}
      tab="verification"
    />
  </PageStory>
)
export const ACROverview = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={acrConnection}
      tab="overview"
      repositoryAccess={<RepositoryAccessDemo connection={acrConnection} />}
    />
  </PageStory>
)
export const ACRVerification = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={verifiedACRConnection}
      tab="verification"
    />
  </PageStory>
)
export const ACRVerifying = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={acrConnection}
      tab="verification"
      isVerifying
    />
  </PageStory>
)
export const ACRTokenFailed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedACRConnection}
      tab="verification"
    />
  </PageStory>
)
export const ACRTrustFailed = () => (
  <PageStory>
    <ConnectionDetail
      {...props}
      connection={failedACRTrustConnection}
      tab="verification"
    />
  </PageStory>
)
