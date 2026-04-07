export default {
  title: 'Sandbox/SandboxRunsSkeleton',
}

import { SandboxRunsSkeleton } from './SandboxRunsSkeleton'

export const Default = () => (
  <div className="flex flex-col gap-1 w-80 p-4">
    <SandboxRunsSkeleton />
  </div>
)

export const Short = () => (
  <div className="flex flex-col gap-1 w-80 p-4">
    <SandboxRunsSkeleton limit={2} />
  </div>
)
