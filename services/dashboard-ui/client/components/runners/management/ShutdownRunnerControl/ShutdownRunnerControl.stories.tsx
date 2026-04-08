export default {
  title: 'Runners/Management/ShutdownRunnerControl',
}

import { ShutdownRunnerControl } from './ShutdownRunnerControl'

export const Unmanaged = () => (
  <div className="p-4">
    <ShutdownRunnerControl runnerId="runner-1" isManaged={false} />
  </div>
)

export const Managed = () => (
  <div className="p-4">
    <ShutdownRunnerControl runnerId="runner-1" isManaged />
  </div>
)
