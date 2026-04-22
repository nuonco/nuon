import { SyncSecretsStepDetails } from './SyncSecretsStepDetails'

export default { title: 'Workflows/SyncSecretsStepDetails' }

export const WithLogStream = () => (
  <SyncSecretsStepDetails
    step={{
      id: 'iws-test-123',
      name: 'sync secrets',
      step_target_type: 'install_workflow_steps',
      step_target_id: 'iws-test-123',
      log_stream: { id: 'log-test-123', open: false },
    } as any}
  />
)

export const WithoutLogStream = () => (
  <SyncSecretsStepDetails
    step={{
      id: 'iws-test-456',
      name: 'sync secrets',
      step_target_type: 'install_workflow_steps',
      step_target_id: 'iws-test-456',
    } as any}
  />
)
