import {
  AppSandboxConfig,
  AppSandboxVariables,
  EmptyStateGraphic,
  Link,
  Text,
} from '@/components'
import { getAppConfigById } from '@/lib'

export const SandboxConfig = async ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfigById({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })
  return config && !error ? (
    <div className="flex flex-col gap-8">
      <AppSandboxConfig sandboxConfig={config?.sandbox} />
      <AppSandboxVariables variables={config?.sandbox?.variables} />
    </div>
  ) : (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyStateGraphic variant="table" />
      <Text className="mt-6" variant="med-14">
        No app sandbox config
      </Text>
      <Text variant="reg-12" className="text-center !inline-block">
        Read more about app sandbox configs{' '}
        <Link
          className="!inline-block"
          href="https://docs.nuon.co/concepts/sandboxes"
          target="_blank"
        >
          here
        </Link>
        .
      </Text>
    </div>
  )
}
