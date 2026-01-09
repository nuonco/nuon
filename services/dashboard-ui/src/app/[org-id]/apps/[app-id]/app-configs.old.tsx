import {
  AppInputConfig,
  AppInputConfigModal,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  EmptyStateGraphic,
  Link,
  Section,
  Text,
} from '@/components'
import { Markdown } from '@/components/common/Showdown'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'

export const InputsConfig = async ({
  appConfigId,
  appId,
  appName,
  orgId,
}: {
  appConfigId: string
  appId: string
  appName: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return config && !error ? (
    <>
      {config?.input && config?.input?.input_groups?.length ? (
        <Section
          className="flex-initial"
          heading="Inputs"
          actions={
            <AppInputConfigModal
              inputConfig={{
                ...config.input,
                input_groups: nestInputsUnderGroups(
                  config.input?.input_groups,
                  config.input?.inputs
                ),
              }}
              appName={appName}
            />
          }
        >
          <AppInputConfig
            inputConfig={{
              ...config.input,
              input_groups: nestInputsUnderGroups(
                config.input?.input_groups,
                config.input?.inputs
              ),
            }}
          />
        </Section>
      ) : null}
    </>
  ) : null
}

export const ReadmeConfig = async ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfig({
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

export const RunnerConfig = async ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfig({
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

export const SandboxConfig = async ({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId: string
  appId: string
  orgId: string
}) => {
  const { data: config, error } = await getAppConfig({
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

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}