import { RunInCloud } from './RunInCloud'
import { VerifyConnection } from './VerifyConnection'
import { ConnectionWizard } from './ConnectionWizard'
import {
  PageStory,
  connection,
  customConnection,
  failedConnection,
  verifiedConnection,
  noop,
} from '../__stories__/fixtures'
export default { title: 'Cloud connections/Wizard' }
export const RunSetup = () => (
  <PageStory>
    <ConnectionWizard step={3} created onStep={noop}>
      <RunInCloud connection={connection} setup={connection.setup} />
    </ConnectionWizard>
  </PageStory>
)
export const CustomSetup = () => (
  <PageStory>
    <ConnectionWizard step={3} created onStep={noop}>
      <RunInCloud
        connection={customConnection}
        setup={customConnection.setup}
      />
    </ConnectionWizard>
  </PageStory>
)
const props = { onVerify: noop, setupHref: '/setup', detailHref: '/detail' }
export const Pending = () => (
  <VerifyConnection {...props} connection={connection} isVerifying={false} />
)
export const Verifying = () => (
  <VerifyConnection {...props} connection={connection} isVerifying />
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
