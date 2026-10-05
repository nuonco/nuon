import { AccountForm } from './AccountForm'
import { ConnectionWizard } from './ConnectionWizard'
import { PageStory, noop } from '../__stories__/fixtures'
export default { title: 'Features / Cloud connections / Wizard / Account' }
export const Empty = () => (
  <PageStory>
    <ConnectionWizard step={1} created={false} onStep={noop}>
      <AccountForm onContinue={noop} />
    </ConnectionWizard>
  </PageStory>
)
export const Filled = () => (
  <PageStory>
    <ConnectionWizard step={1} created={false} onStep={noop}>
      <AccountForm
        onContinue={noop}
        values={{
          name: 'acme-production',
          target_id: '123456789012',
          role_name: 'acme-nuon',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
