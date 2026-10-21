import { useStore } from '@tanstack/react-form'
import { Card } from '@/components/common/Card'
import { CloudRegion } from '@/components/common/CloudRegion'
import { Code } from '@/components/common/Code'
import { Expand } from '@/components/common/Expand'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import type { ColumnDef } from '@tanstack/react-table'
import type { TGroupStepSelection } from '@/components/installs/CreateInstall/GroupStep'
import {
  getEditableInputs,
  isBooleanInput,
  type InstallFormApi,
  type InstallPlatform,
} from '@/components/installs/forms/InstallForm'
import type {
  TAction,
  TAPIError,
  TAppInputConfig,
  TCloudConnectionSummary,
  TComponent,
} from '@/types'
import {
  buildDeploymentPlan,
  type IDeploymentPlanRow,
} from '../deployment-plan'
import { StepActions } from '../StepActions'

interface IReviewStep {
  form: InstallFormApi
  platform?: InstallPlatform
  appName: string
  branchName: string
  selection: TGroupStepSelection
  inputConfig?: TAppInputConfig
  actions: TAction[]
  components: TComponent[]
  cloudConnections?: TCloudConnectionSummary[]
  submitError?: TAPIError | null
  submitting: boolean
  canSubmit: boolean
  disabledReason?: string
  onAdvance: () => void
  onGoBack?: () => void
  cancelHref?: string
}

const displayValue = (value: string | boolean | undefined) => {
  if (typeof value === 'boolean') return value ? 'On' : 'Off'
  if (typeof value === 'string' && value.trim()) return value
  return 'Not set'
}

const ActionNames = ({ actions }: { actions: TAction[] }) =>
  actions.length > 0 ? (
    <span className="flex flex-wrap gap-1">
      {actions.map((action) => (
        <Code key={action.id} variant="inline">
          {action.name}
        </Code>
      ))}
    </span>
  ) : (
    <Text variant="subtext" theme="neutral">
      None
    </Text>
  )

const deploymentColumns: ColumnDef<IDeploymentPlanRow>[] = [
  {
    id: 'component',
    header: 'Component',
    cell: ({ row }) => (
      <span className="flex flex-col gap-1">
        <Code variant="inline">{row.original.component.name}</Code>
        <Text variant="subtext" theme="neutral">
          {row.original.component.type}
        </Text>
      </span>
    ),
  },
  {
    id: 'before',
    header: 'Before deploy',
    cell: ({ row }) => <ActionNames actions={row.original.before} />,
    enableSorting: false,
  },
  {
    id: 'after',
    header: 'After deploy',
    cell: ({ row }) => <ActionNames actions={row.original.after} />,
    enableSorting: false,
  },
]

export const ReviewStep = ({
  form,
  platform,
  appName,
  branchName,
  selection,
  inputConfig,
  actions,
  components,
  cloudConnections,
  submitError,
  submitting,
  canSubmit,
  disabledReason,
  onAdvance,
  onGoBack,
  cancelHref,
}: IReviewStep) => {
  const values = useStore(form.store, (state) => state.values)
  const labels = values.labels.filter((label) => label.key.trim())
  const connection = cloudConnections?.find(
    (item) => item.id === values.cloud_connection_id
  )
  const inputs = getEditableInputs(inputConfig).filter((input) => input?.name)
  const deploymentPlan = buildDeploymentPlan(components, actions)
  const region =
    platform === 'azure' ? values.location : platform === 'gcp' ? values.gcp_project_id : values.region

  return (
    <div className="flex flex-col gap-8">
      <FormErrorBanner error={submitError} fallback="Unable to create install" />
      <Card elevation="1">
        <div className="grid gap-5 sm:grid-cols-2">
          <LabeledValue label="App">{appName}</LabeledValue>
          <LabeledValue label="Branch">
            <Code variant="inline">{branchName}</Code>
          </LabeledValue>
          <LabeledValue label="Install group">
            {selection.group?.name ?? 'None'}
          </LabeledValue>
          <LabeledValue label="Install name">
            <Code variant="inline">{values.name.trim() || 'None'}</Code>
          </LabeledValue>
          {platform === 'aws' || platform === 'azure' ? (
            <LabeledValue label={platform === 'azure' ? 'Azure location' : 'AWS region'}>
              {region ? (
                <CloudRegion
                  platform={platform}
                  region={platform === 'aws' ? values.region : undefined}
                  location={platform === 'azure' ? values.location : undefined}
                />
              ) : (
                'None'
              )}
            </LabeledValue>
          ) : null}
          {platform === 'aws' ? (
            <LabeledValue label="AWS account ID">
              {values.aws_account_id ? (
                <Code variant="inline">{values.aws_account_id}</Code>
              ) : (
                'None'
              )}
            </LabeledValue>
          ) : null}
          {platform === 'azure' ? (
            <LabeledValue label="Azure subscription ID">
              {values.azure_subscription_id ? (
                <Code variant="inline">{values.azure_subscription_id}</Code>
              ) : (
                'None'
              )}
            </LabeledValue>
          ) : null}
          {platform === 'gcp' ? (
            <LabeledValue label="GCP project ID">
              {values.gcp_project_id ? (
                <Code variant="inline">{values.gcp_project_id}</Code>
              ) : (
                'None'
              )}
            </LabeledValue>
          ) : null}
          {connection ? (
            <LabeledValue label="Cloud connection">{connection.name}</LabeledValue>
          ) : null}
          <LabeledValue label="Auto-approve changes">
            {values.autoApprove ? 'On' : 'Off'}
          </LabeledValue>
          <LabeledValue label="Initial provisioning">
            {values.stackOnly ? 'Stack and runner only' : 'Full install'}
          </LabeledValue>
        </div>
        <Expand
          id="create-review-labels"
          heading={`Labels (${labels.length})`}
          className="rounded-md border"
          headerClassName="px-3"
        >
          <div className="flex flex-wrap gap-2 border-t p-3">
            {labels.length > 0 ? (
              labels.map((label) => (
                <LabelBadge
                  key={`${label.key}=${label.value}`}
                  labelKey={label.key.trim()}
                  labelValue={label.value.trim()}
                />
              ))
            ) : (
              <Text variant="subtext" theme="neutral">
                No labels
              </Text>
            )}
          </div>
        </Expand>
        <Expand
          id="create-review-inputs"
          heading={`Inputs (${inputs.length})`}
          className="rounded-md border"
          headerClassName="px-3"
        >
          <div className="grid gap-4 border-t p-3 sm:grid-cols-2">
            {inputs.length > 0 ? (
              inputs.map((input) => (
                <LabeledValue
                  key={input.name}
                  label={input.display_name || input.name || 'Input'}
                >
                  {isBooleanInput(input) ||
                  typeof values.inputs[input.name!] === 'string'
                    ? displayValue(values.inputs[input.name!])
                    : 'Not set'}
                </LabeledValue>
              ))
            ) : (
              <Text variant="subtext" theme="neutral">
                No inputs
              </Text>
            )}
          </div>
        </Expand>
        {values.stackOnly ? null : (
          <div className="flex flex-col gap-3 border-t pt-5">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Component deployment plan</Text>
              <Text variant="subtext" theme="neutral">
                Pre- and post-deploy actions run around their component.
              </Text>
            </div>
            <Table
              columns={deploymentColumns}
              data={deploymentPlan}
              enableSearch={false}
              enableSorting={false}
              emptyStateProps={{
                emptyTitle: 'No components',
                emptyMessage: 'This app config does not deploy components.',
              }}
            />
          </div>
        )}
        <div className="flex flex-col gap-2 border-t pt-5">
          <Text weight="strong">What will happen</Text>
          <Text theme="neutral">
            Nuon will create the install, provision its stack and runner
            {values.stackOnly
              ? ', and stop before provisioning the sandbox and components.'
              : ', then provision the sandbox and deploy the app components.'}
          </Text>
        </div>
      </Card>
      <StepActions
        onGoBack={onGoBack}
        onAdvance={onAdvance}
        label={submitting ? 'Creating install...' : 'Create install'}
        cancelHref={cancelHref}
        disabled={submitting || !canSubmit}
        disabledReason={
          submitting ? undefined : disabledReason || 'Complete the required fields'
        }
      />
    </div>
  )
}
