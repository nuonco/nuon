import { Card, Heading, Text, Link } from '@/components'
import {
  getAppInputLatestConfig,
  getAppRunnerLatestConfig,
  getAppSandboxLatestConfig,
} from '@/lib'

export default async function App({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [inputCfg, runnerCfg, sandboxCfg] = await Promise.all([
    getAppInputLatestConfig({ appId, orgId }),
    getAppRunnerLatestConfig({ appId, orgId }),
    getAppSandboxLatestConfig({ appId, orgId }),
  ])

  return (
    <div className="flex flex-col md:flex-row gap-6">
      <div className="flex-auto flex flex-col gap-6">
        <Card>
          <Heading>Sandbox config</Heading>
          <div>{sandboxCfg.cloud_platform}</div>
        </Card>

        <Card>
          <Heading>Inputs</Heading>
          <div className="flex flex-col gap-4">
            {inputCfg.inputs.map((input) => (
              <span key={input.id} className="flex items-center gap-4 text-sm">
                {input.display_name}: {input.default || 'No default'}
              </span>
            ))}
          </div>
        </Card>
      </div>

      <div className="flex-auto flex flex-col gap-6">
        <Card>
          <Heading>Runner config</Heading>
          <div>{runnerCfg.app_runner_type}</div>
        </Card>
      </div>
    </div>
  )
}
