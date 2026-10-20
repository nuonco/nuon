import { useState } from 'react'
import { useStore } from '@tanstack/react-form'
import { Expand } from '@/components/common/Expand'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Text } from '@/components/common/Text'
import { FieldRow } from '@/components/installs/forms/InstallForm/FieldRow'
import { InstallPlatformFields } from '@/components/installs/forms/InstallForm/InstallPlatformFields'
import { StackOnlyCheckbox } from '@/components/installs/forms/InstallForm/StackOnlyCheckbox'
import { LabelsField } from '@/components/installs/forms/InstallForm/LabelsField'
import type { TGroupStepSelection } from '@/components/installs/CreateInstall/GroupStep'
import type { InstallFormApi } from '@/components/installs/forms/InstallForm'
import type { InstallPlatform } from '@/components/installs/forms/InstallForm'
import type { TAppBranchConfig, TCloudConnectionSummary } from '@/types'
import { assignmentBlocker, installDetailsBlocker } from '../readiness'
import { StepActions } from '../StepActions'

const NAME_CHECK_DEBOUNCE_MS = 400

interface IDetailsStep {
  form: InstallFormApi
  platform?: InstallPlatform
  cloudConnections?: TCloudConnectionSummary[]
  groupConfig?: TAppBranchConfig
  selection: TGroupStepSelection
  requireTargetAccount: boolean
  validateName?: (name: string) => Promise<string | undefined>
  onAdvance: () => void
  onGoBack?: () => void
  cancelHref?: string
}

export const DetailsStep = ({
  form,
  platform,
  cloudConnections,
  groupConfig,
  selection,
  requireTargetAccount,
  validateName,
  onAdvance,
  onGoBack,
  cancelHref,
}: IDetailsStep) => {
  const values = useStore(form.store, (state) => state.values)
  const isValidating = useStore(form.store, (state) => state.isFieldsValidating)
  const [checkingName, setCheckingName] = useState(false)
  const blocker =
    installDetailsBlocker({
      values,
      platform,
      requireTargetAccount,
    }) ?? assignmentBlocker(selection, values.labels, groupConfig)

  const continueDetails = async () => {
    setCheckingName(true)
    const errors = await form.validateField('name', 'submit')
    setCheckingName(false)
    if (errors?.length) return
    onAdvance()
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-5">
        <FieldRow
          labelText="Install name"
          required
          helpText="A unique name for this install"
        >
          <form.Field
            name="name"
            validators={
              validateName
                ? {
                    onChangeAsyncDebounceMs: NAME_CHECK_DEBOUNCE_MS,
                    onChangeAsync: ({ value }) => validateName(value),
                    onBlurAsync: ({ value }) => validateName(value),
                    onSubmitAsync: ({ value }) => validateName(value),
                  }
                : undefined
            }
          >
            {(field) => (
              <FormInput field={field} placeholder="Enter install name" />
            )}
          </form.Field>
        </FieldRow>

        {platform ? (
          <InstallPlatformFields
            form={form}
            platform={platform}
            cloudConnections={cloudConnections}
            requireTargetAccount={requireTargetAccount}
          />
        ) : null}

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
                        Automatically approve and apply all future changes
                        without manual confirmation. You can change this later
                        in the install settings.
                      </Text>
                    </div>
                  ),
                }}
              />
            )}
          </form.Field>
        </FieldRow>

        <StackOnlyCheckbox form={form} />

        <LabelsField form={form} isOpen />

        {platform === 'aws' ? (
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
        ) : null}
      </div>
      <StepActions
        onGoBack={onGoBack}
        onAdvance={continueDetails}
        label="Continue"
        showCaret
        cancelHref={cancelHref}
        disabled={!!blocker || isValidating || checkingName}
        disabledReason={
          blocker ??
          (isValidating || checkingName ? 'Checking the install name' : undefined)
        }
      />
    </div>
  )
}
