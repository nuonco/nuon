import { Banner } from './Banner'
import { Button } from './Button'
import { Text } from './Text'

export const LongContent = () => (
  <Banner theme="warn">
    <div className="flex items-center justify-between gap-4">
      <div className="flex flex-col">
        <Text weight="strong">Component needs approval</Text>
        <Text variant="subtext" theme="neutral">
          Review the components changes and approve for deployment.
        </Text>
      </div>
      <div className="flex gap-4">
        <Button variant="danger">Deny</Button>
        <Button variant="primary">Approve</Button>
      </div>
    </div>
  </Banner>
)

export const AllThemes = () => (
  <div className="space-y-4">
    <Banner theme="error">
      Error: Something went wrong. Please try again.
    </Banner>
    <Banner theme="warn">Warning: This action cannot be undone.</Banner>
    <Banner theme="info">
      Info: Your changes have been saved automatically.
    </Banner>
    <Banner theme="success">
      Success: Your deployment was completed successfully!
    </Banner>
    <Banner theme="default">
      Default: This is a default banner with important information.
    </Banner>
  </div>
)

export const SimpleMessages = () => (
  <div className="space-y-4">
    <Banner theme="error">Something went wrong</Banner>
    <Banner theme="warn">Caution required</Banner>
    <Banner theme="info">Good to know</Banner>
    <Banner theme="success">All done!</Banner>
  </div>
)

export const WithComplexContent = () => (
  <div className="space-y-4">
    <Banner theme="info">
      <div className="space-y-2">
        <Text weight="strong">Update Available</Text>
        <Text variant="subtext">
          A new version of the application is available. Update now to get the
          latest features and security improvements.
        </Text>
      </div>
    </Banner>
    <Banner theme="success">
      <div className="flex items-start justify-between">
        <div>
          <Text weight="strong">Deploy Successful</Text>
          <Text variant="subtext">
            Your application has been deployed to production successfully.
          </Text>
        </div>
        <Button size="sm">View Details</Button>
      </div>
    </Banner>
  </div>
)
