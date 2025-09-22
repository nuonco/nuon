import { EmptyStateGraphic, Section, Text, Markdown } from '@/components'
import { getAppConfigById } from '@/lib'

export const ReadmeConfig = async ({
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
  })
  return config && !error ? (
    <Section className="border-r overflow-x-auto" heading="README">
      <Markdown content={config?.readme} />
    </Section>
  ) : (
    <Section className="border-r overflow-x-auto" heading="README">
      <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
        <EmptyStateGraphic variant="table" />
        <Text className="mt-6" variant="med-14">
          No README in app config
        </Text>
        <Text variant="reg-12" className="text-center !inline-block">
          You can add a README for your app in your app config TOML file.
        </Text>
      </div>
    </Section>
  )
}
