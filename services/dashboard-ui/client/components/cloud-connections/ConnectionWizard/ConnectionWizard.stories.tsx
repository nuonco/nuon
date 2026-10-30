import { RunInCloud } from './RunInCloud'
import { VerifyConnection } from './VerifyConnection'
import { ConnectionWizard } from './ConnectionWizard'
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
  initialECRConnection,
  initialGARConnection,
  initialACRConnection,
  failedGARConnection,
  failedGARTrustConnection,
  verifiedGARConnection,
  failedECRConnection,
  failedConnection,
  verifiedECRConnection,
  verifiedConnection,
  noop,
} from '../__stories__/fixtures'
export default { title: 'Features / Cloud connections / Wizard' }
export const RunSetup = () => (
  <PageStory>
    <ConnectionWizard connection={connection}>
      <RunInCloud connection={connection} setup={connection.setup} />
    </ConnectionWizard>
  </PageStory>
)
export const CustomSetup = () => (
  <PageStory>
    <ConnectionWizard connection={customConnection}>
      <RunInCloud
        connection={customConnection}
        setup={customConnection.setup}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ECRSetup = () => (
  <PageStory>
    <ConnectionWizard connection={ecrConnection}>
      <RunInCloud
        connection={ecrConnection}
        setup={initialECRConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={ecrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ECRResumeSetup = () => (
  <PageStory>
    <ConnectionWizard connection={failedECRConnection}>
      <RunInCloud
        connection={failedECRConnection}
        setup={failedECRConnection.setup}
        repositoryAccess={<RepositoryAccessDemo connection={ecrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ECRTrustOnlySetup = () => (
  <PageStory>
    <ConnectionWizard connection={ecrConnection}>
      <RunInCloud
        connection={ecrConnection}
        setup={ecrConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={ecrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
const props = { onVerify: noop, setupHref: '/setup', detailHref: '/detail' }
export const ECRPending = () => (
  <VerifyConnection {...props} connection={ecrConnection} isVerifying={false} />
)
export const ECRVerifying = () => (
  <VerifyConnection {...props} connection={ecrConnection} isVerifying />
)
export const ECRVerified = () => (
  <VerifyConnection
    {...props}
    connection={verifiedECRConnection}
    isVerifying={false}
  />
)
export const ECRTokenFailed = () => (
  <VerifyConnection
    {...props}
    connection={failedECRConnection}
    isVerifying={false}
  />
)
export const ECRTrustFailed = () => (
  <VerifyConnection
    {...props}
    connection={{
      ...ecrConnection,
      status: 'error',
      status_message: failedConnection.status_message,
    }}
    isVerifying={false}
  />
)
export const ECRRequestFailed = () => (
  <VerifyConnection
    {...props}
    connection={ecrConnection}
    isVerifying={false}
    error={{
      error: 'ECR token retrieval failed',
      description: 'AWS denied ecr:GetAuthorizationToken for this role.',
      user_error: true,
    }}
  />
)
export const Pending = () => (
  <VerifyConnection {...props} connection={connection} isVerifying={false} />
)
export const Verifying = () => (
  <VerifyConnection {...props} connection={connection} isVerifying />
)
export const StillChecking = () => (
  <VerifyConnection
    {...props}
    connection={failedConnection}
    isVerifying={false}
    verificationTimedOut
  />
)
export const Verified = () => (
  <VerifyConnection
    {...props}
    connection={verifiedConnection}
    isVerifying={false}
  />
)
export const FailedTrust = () => (
  <VerifyConnection
    {...props}
    connection={failedConnection}
    isVerifying={false}
  />
)
export const FailedPermissions = () => (
  <VerifyConnection
    {...props}
    connection={{
      ...failedConnection,
      status_message: 'AWS denied cloudformation:DescribeStacks.',
    }}
    isVerifying={false}
  />
)
export const RequestFailed = () => (
  <VerifyConnection
    {...props}
    connection={verifiedConnection}
    isVerifying={false}
    error={{
      error: 'Unable to exchange the OIDC token.',
      description: 'Unable to exchange the OIDC token.',
      user_error: false,
    }}
  />
)
export const GARSetup = () => (
  <PageStory>
    <ConnectionWizard connection={garConnection}>
      <RunInCloud
        connection={garConnection}
        setup={initialGARConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={garConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const GARResumeSetup = () => (
  <PageStory>
    <ConnectionWizard connection={failedGARConnection}>
      <RunInCloud
        connection={failedGARConnection}
        setup={failedGARConnection.setup}
        repositoryAccess={<RepositoryAccessDemo connection={garConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const GARTrustOnlySetup = () => (
  <PageStory>
    <ConnectionWizard connection={garConnection}>
      <RunInCloud
        connection={garConnection}
        setup={garConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={garConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const GARPending = () => (
  <VerifyConnection {...props} connection={garConnection} isVerifying={false} />
)
export const GARVerifying = () => (
  <VerifyConnection {...props} connection={garConnection} isVerifying />
)
export const GARVerified = () => (
  <VerifyConnection
    {...props}
    connection={verifiedGARConnection}
    isVerifying={false}
  />
)
export const GARImpersonationFailed = () => (
  <VerifyConnection
    {...props}
    connection={failedGARConnection}
    isVerifying={false}
  />
)
export const GARTrustFailed = () => (
  <VerifyConnection
    {...props}
    connection={failedGARTrustConnection}
    isVerifying={false}
  />
)
export const GARStillChecking = () => (
  <VerifyConnection
    {...props}
    connection={garConnection}
    isVerifying={false}
    verificationTimedOut
  />
)
export const GARRequestFailed = () => (
  <VerifyConnection
    {...props}
    connection={garConnection}
    isVerifying={false}
    error={{
      error: 'Federation trust check failed',
      description:
        'GCP denied service account impersonation for this connection subject.',
      user_error: true,
    }}
  />
)
export const ACRSetup = () => (
  <PageStory>
    <ConnectionWizard connection={acrConnection}>
      <RunInCloud
        connection={acrConnection}
        setup={initialACRConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={acrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ACRResumeSetup = () => (
  <PageStory>
    <ConnectionWizard connection={failedACRConnection}>
      <RunInCloud
        connection={failedACRConnection}
        setup={failedACRConnection.setup}
        repositoryAccess={<RepositoryAccessDemo connection={acrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ACRTrustOnlySetup = () => (
  <PageStory>
    <ConnectionWizard connection={acrConnection}>
      <RunInCloud
        connection={acrConnection}
        setup={acrConnection.setup}
        hasInitialSetup
        repositoryAccess={<RepositoryAccessDemo connection={acrConnection} />}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ACRPending = () => (
  <VerifyConnection {...props} connection={acrConnection} isVerifying={false} />
)
export const ACRVerifying = () => (
  <VerifyConnection {...props} connection={acrConnection} isVerifying />
)
export const ACRVerified = () => (
  <VerifyConnection
    {...props}
    connection={verifiedACRConnection}
    isVerifying={false}
  />
)
export const ACRTokenExchangeFailed = () => (
  <VerifyConnection
    {...props}
    connection={failedACRConnection}
    isVerifying={false}
  />
)
export const ACRTrustFailed = () => (
  <VerifyConnection
    {...props}
    connection={failedACRTrustConnection}
    isVerifying={false}
  />
)
export const ACRTokenFailed = () => (
  <VerifyConnection
    {...props}
    connection={{
      ...acrConnection,
      status: 'error',
      status_message: 'ACR token exchange failed: registry access denied.',
    }}
    isVerifying={false}
  />
)
export const ACRStillChecking = () => (
  <VerifyConnection
    {...props}
    connection={acrConnection}
    isVerifying={false}
    verificationTimedOut
  />
)
export const ACRRequestFailed = () => (
  <VerifyConnection
    {...props}
    connection={acrConnection}
    isVerifying={false}
    error={{
      error: 'ACR token exchange failed',
      description: 'Registry token exchange was denied.',
      user_error: true,
    }}
  />
)
