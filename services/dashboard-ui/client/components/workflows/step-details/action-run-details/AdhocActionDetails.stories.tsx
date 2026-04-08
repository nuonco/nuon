export default {
  title: 'Workflows/StepDetails/AdhocActionDetails',
}

import { AdhocActionDetails } from './AdhocActionDetails'
import type { TInstallActionRun } from '@/types'

const mockActionRun: TInstallActionRun = {
  id: 'run-adhoc-1',
  steps: [
    {
      id: 'step-1',
      status: 'finished',
      execution_duration: 8400000000,
      adhoc_config: {
        name: 'restart-service',
        command: 'systemctl restart my-service',
        env_vars: {
          SERVICE_NAME: 'my-service',
          RESTART_TIMEOUT: '30',
        },
      },
    },
  ],
  config: null,
} as TInstallActionRun

export const Default = () => <AdhocActionDetails actionRun={mockActionRun} />

export const WithInlineContents = () => (
  <AdhocActionDetails
    actionRun={{
      ...mockActionRun,
      steps: [
        {
          ...mockActionRun.steps?.[0],
          adhoc_config: {
            name: 'custom-script',
            inline_contents: '#!/bin/bash\necho "Running custom script"\nls -la /var/log',
            env_vars: {},
          },
        },
      ],
    } as TInstallActionRun}
  />
)
