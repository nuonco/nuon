import { useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { InstallView } from '@/views/install/InstallView'
import {
  refreshInstallFixture,
  viewPath,
} from '@/views/install/install-fixture'
import {
  runnerDemandFixture,
  type TRunnerDemandState,
} from './RunnerDemand.fixtures'

export default {
  title: 'Playground / Installs / On demand runner',
  meta: { fullBleed: true, installViews: true },
}

const RunnerDemandPreview = ({
  initialState,
}: {
  initialState: TRunnerDemandState
}) => {
  const client = useQueryClient()
  const [state, setState] = useState(initialState)
  const current = useRef(state)
  const transition = (next: TRunnerDemandState) => {
    current.current = next
    setState(next)
    refreshInstallFixture()
    client.invalidateQueries()
  }
  const [fixture] = useState(() =>
    runnerDemandFixture(
      () => current.current,
      () => transition('cancelled')
    )
  )

  return (
    <div className="runner-demand-preview flex h-screen flex-col">
      <style>{`
        body:has(.runner-demand-preview) :is(.panel-wrapper, .panel) {
          top: calc(var(--spacing) * 12);
          height: calc(100vh - var(--spacing) * 12);
        }
      `}</style>
      <div className="relative z-20 flex h-12 shrink-0 items-center gap-3 border-b bg-elevation-1 px-4">
        <Text variant="subtext" theme="neutral">
          Ladle simulation
        </Text>
        <span className="hidden md:inline">
          <Text variant="subtext">
            {state === 'waiting'
              ? 'Runner offline · job queued for 2 days'
              : state === 'running'
                ? 'Runner online · job executing'
                : state === 'completed'
                  ? 'Job completed'
                  : 'Cancelled before pickup'}
          </Text>
        </span>
        {state === 'waiting' ? (
          <Button size="sm" onClick={() => transition('running')}>
            Simulate runner pickup
          </Button>
        ) : state === 'running' ? (
          <Button size="sm" onClick={() => transition('completed')}>
            Simulate completion
          </Button>
        ) : (
          <Button size="sm" onClick={() => transition('waiting')}>
            Reset simulation
          </Button>
        )}
      </div>
      <div className="min-h-0 flex-1 [&>div]:h-full">
        <InstallView
          fixture={fixture}
          path={viewPath('/deployments/wf-deploy-1/workflow')}
        />
      </div>
    </div>
  )
}

export const WaitingForRunner = () => (
  <RunnerDemandPreview initialState="waiting" />
)
export const RunnerPickedUp = () => (
  <RunnerDemandPreview initialState="running" />
)
export const Completed = () => <RunnerDemandPreview initialState="completed" />
export const CancelledBeforePickup = () => (
  <RunnerDemandPreview initialState="cancelled" />
)
