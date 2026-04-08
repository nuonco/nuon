export default {
  title: 'Runners/CancelRunnerJob',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CancelRunnerJobModal } from './CancelRunnerJob'
import type { TRunnerJob } from '@/types'

const mockJob = { id: 'job_abc123' } as TRunnerJob
const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CancelRunnerJobModal
      runnerJob={mockJob}
      jobType="build"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CancelRunnerJobModal
      runnerJob={mockJob}
      jobType="deploy"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CancelRunnerJobModal
      runnerJob={mockJob}
      jobType="build"
      isPending={false}
      error={{ error: 'Failed to cancel job: runner not reachable' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
