import { RunLogs, RunOutputs, RunSummary, RunTrace } from './RunDetail'

export default {
  title: 'Playground/Lite/RunDetail',
}

export const Summary = () => (
  <div className="p-4">
    <RunSummary />
  </div>
)

export const Logs = () => (
  <div className="p-4">
    <RunLogs />
  </div>
)

export const Trace = () => (
  <div className="p-4">
    <RunTrace />
  </div>
)

export const Outputs = () => (
  <div className="p-4">
    <RunOutputs />
  </div>
)
