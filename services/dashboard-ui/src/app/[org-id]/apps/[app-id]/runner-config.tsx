import { AppRunnerConfig, EmptyStateGraphic, Link, Text } from '@/components'
import { getAppConfigById } from '@/lib'

export const RunnerConfig = async ({
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
    <AppRunnerConfig runnerConfig={config?.runner} />
  ) : (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyStateGraphic variant="table" />
      <Text className="mt-6" variant="med-14">
        No app runner config
      </Text>
      <Text variant="reg-12" className="text-center !inline-block">
        Read more about app runner configs{' '}
        <Link
          className="!inline-block"
          href="https://docs.nuon.co/concepts/runners"
          target="_blank"
        >
          here
        </Link>
        .
      </Text>
    </div>
  )
}
