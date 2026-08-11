import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Text } from '@/components/common/Text'
import { Expand } from '@/components/common/Expand'
import { RoleSelector } from '@/components/roles/RoleSelector'
import type { TAppInputConfig, TAWSAccountConnection, TInstall } from '@/types'
import { FieldRow } from './FieldRow'
import { InstallInputFields } from './InstallInputFields'
import { InstallPlatformFields } from './InstallPlatformFields'
import { LabelsField } from './LabelsField'
import type { InstallFormApi } from './useInstallForm'
import type { InstallFormMode, InstallPlatform } from './schema'

export interface IInstallForm {
  form: InstallFormApi
  mode: InstallFormMode
  platform?: InstallPlatform
  inputConfig?: TAppInputConfig
  install?: TInstall
  installId?: string
  awsAccountConnections?: TAWSAccountConnection[]
  requireTargetAccount?: boolean
  autoApproveDescription?: string
  showNameField?: boolean
}

export const InstallForm = ({
  form,
  mode,
  platform,
  inputConfig,
  install,
  installId,
  awsAccountConnections,
  requireTargetAccount,
  autoApproveDescription,
  showNameField = true,
}: IInstallForm) => {
  const showName = mode === 'create' || showNameField

  return (
    <form
      autoComplete="off"
      noValidate
      onSubmit={(e) => e.preventDefault()}
      className="flex flex-col gap-8 max-w-4xl pb-4"
    >
      {showName && (
        <FieldRow
          labelText="Install name"
          required
          helpText="A unique name for this install"
        >
          <form.Field name="name">
            {(field) => (
              <FormInput field={field} placeholder="Enter install name" />
            )}
          </form.Field>
        </FieldRow>
      )}

      {mode === 'create' && platform && (
        <InstallPlatformFields
          form={form}
          platform={platform}
          awsAccountConnections={awsAccountConnections}
          requireTargetAccount={requireTargetAccount}
        />
      )}

      {mode === 'create' && (
        <FieldRow
          labelText="Deployment approval"
          helpText="Choose how deployments should be approved"
        >
          <form.Field name="autoApprove">
            {(field) => (
              <FormCheckbox
                field={field}
                className="mt-[6px]"
                labelProps={{
                  className: 'items-start',
                  labelText: (
                    <div className="flex flex-col gap-1">
                      <Text variant="body" weight="stronger">
                        Auto-approve changes
                      </Text>
                      <Text variant="subtext" theme="neutral">
                        {autoApproveDescription ??
                          'Automatically approve and apply all future changes without manual confirmation. You can change this later in the install settings.'}
                      </Text>
                    </div>
                  ),
                }}
              />
            )}
          </form.Field>
        </FieldRow>
      )}

      {mode === 'create' && <LabelsField form={form} />}

      {mode === 'create' && platform === 'aws' && (
        <Expand
          id="advanced-stack-overrides"
          heading="Advanced"
          headerClassName="!px-4 bg-code"
          className="mt-2 border rounded-md"
        >
          <div className="flex flex-col gap-6 p-4 border-t">
            <FieldRow
              labelText="VPC template URL override"
              optional
              helpText="Override the app-level VPC nested CloudFormation template"
            >
              <form.Field name="vpcTemplateUrl">
                {(field) => (
                  <FormInput
                    field={field}
                    placeholder="https://s3.amazonaws.com/..."
                  />
                )}
              </form.Field>
            </FieldRow>

            <FieldRow
              labelText="Runner template URL override"
              optional
              helpText="Override the app-level runner nested CloudFormation template"
            >
              <form.Field name="runnerTemplateUrl">
                {(field) => (
                  <FormInput
                    field={field}
                    placeholder="https://s3.amazonaws.com/..."
                  />
                )}
              </form.Field>
            </FieldRow>
          </div>
        </Expand>
      )}

      <InstallInputFields
        form={form}
        inputConfig={inputConfig}
        install={install}
      />

      {mode === 'edit' && (
        <form.Field name="role">
          {(field) => (
            <RoleSelector
              installId={installId ?? ''}
              value={field.state.value}
              onChange={field.handleChange}
              name="role"
            />
          )}
        </form.Field>
      )}
    </form>
  )
}
