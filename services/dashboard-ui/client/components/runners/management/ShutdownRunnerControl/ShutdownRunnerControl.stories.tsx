export default {
  title: 'Features / Runners / Management / Shutdown runner control',
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
