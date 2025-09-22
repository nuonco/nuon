import { AppInputConfig, AppInputConfigModal, Section } from '@/components'
import { getAppConfigById } from '@/lib'
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
  const { data: config, error } = await getAppConfigById({
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

function nestInputsUnderGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
