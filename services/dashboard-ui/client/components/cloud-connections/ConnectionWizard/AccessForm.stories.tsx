import { AccessForm } from './AccessForm'
import { ConnectionWizard } from './ConnectionWizard'
import { PageStory, noop } from '../__stories__/fixtures'
export default { title: 'Cloud connections/Wizard/Access' }
export const Choose = () => (
  <PageStory>
    <ConnectionWizard step={2} created={false} onStep={noop}>
      <AccessForm onContinue={noop} onBack={noop} />
    </ConnectionWizard>
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ConnectionWizard step={2} created={false} onStep={noop}>
      <AccessForm
        onContinue={noop}
        onBack={noop}
        error={{
          error: 'Connection creation failed',
          description: 'This role already has a cloud connection.',
          user_error: true,
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
