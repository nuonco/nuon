export default {
  title: 'Features / Installs / Create wizard',
}

import { useEffect, useState } from 'react'
import {
  labelsFromGroupSelection,
  type TGroupStepSelection,
} from '@/components/installs/CreateInstall/GroupStep'
import { useInstallForm } from '@/components/installs/forms/InstallForm'
import type {
  TAction,
  TAppBranchConfig,
  TAppInputConfig,
  TComponent,
} from '@/types'
import { CreateInstallWizard } from './CreateInstallWizard'
import { AppStep } from './steps/AppStep'
import { ReviewStep } from './steps/ReviewStep'
import { CREATE_WIZARD_STEPS } from './wizard-path'

const config = {
  id: 'cfg-1',
  install_groups: [
    {
      id: 'g-prod',
      name: 'Production',
      label_selector: { match_labels: { env: 'production' } },
    },
    {
      id: 'g-default',
      name: 'Remaining installs',
      default: true,
    },
  ],
} as unknown as TAppBranchConfig

export const SelectApp = () => {
  const [selected, setSelected] = useState<TGroupStepSelection>({
    mode: 'none',
    group: null,
  })
  const [labels, setLabels] = useState<{ key: string; value: string }[]>([])

  return (
    <div className="bg-background p-6">
      <CreateInstallWizard
        steps={CREATE_WIZARD_STEPS.map((step, index) => ({
          id: step.id,
          label: step.navLabel,
          status: index === 0 ? 'current' : 'upcoming',
        }))}
        selectedId="app"
        title={CREATE_WIZARD_STEPS[0].title}
        description={CREATE_WIZARD_STEPS[0].description}
        onSelect={() => {}}
      >
        <AppStep
          appName="payments"
          branchName="main"
          config={config}
          selected={selected}
          labels={labels}
          onSelect={(next) => {
            setSelected(next)
            setLabels(
              Object.entries(labelsFromGroupSelection(next)).map(
                ([key, value]) => ({ key, value })
              )
            )
          }}
          onAdvance={() => {}}
        />
      </CreateInstallWizard>
    </div>
  )
}

const reviewInputConfig = {
  id: 'input-config-1',
  input_groups: [
    {
      id: 'input-group-1',
      name: 'application',
      display_name: 'Application',
      app_inputs: [
        {
          id: 'input-1',
          name: 'environment',
          display_name: 'Environment',
          source: 'vendor',
          default: 'production',
        },
      ],
    },
  ],
} as TAppInputConfig

const reviewComponents = [
  { id: 'component-api', name: 'api', type: 'helm_chart' },
  { id: 'component-worker', name: 'worker', type: 'terraform_module' },
] as TComponent[]

const reviewActions = [
  {
    id: 'action-migrate',
    name: 'migrate-database',
    configs: [
      {
        triggers: [
          {
            id: 'trigger-1',
            type: 'pre-deploy-component',
            component_id: 'component-api',
          },
        ],
      },
    ],
  },
  {
    id: 'action-notify',
    name: 'notify-release',
    configs: [
      {
        triggers: [
          {
            id: 'trigger-2',
            type: 'post-deploy-all-components',
          },
        ],
      },
    ],
  },
] as TAction[]

export const Review = () => {
  const { form } = useInstallForm({
    mode: 'create',
    platform: 'aws',
    inputConfig: reviewInputConfig,
    defaultAutoApprove: true,
    onSubmit: () => {},
  })
  useEffect(() => {
    form.setFieldValue('name', 'acme-production')
    form.setFieldValue('region', 'us-east-1')
    form.setFieldValue('labels', [
      { key: 'environment', value: 'production' },
    ])
  }, [form])

  return (
    <div className="bg-background p-6">
      <CreateInstallWizard
        steps={CREATE_WIZARD_STEPS.map((step, index) => ({
          id: step.id,
          label: step.navLabel,
          status: index < 3 ? 'done' : index === 3 ? 'current' : 'upcoming',
        }))}
        selectedId="review"
        title={CREATE_WIZARD_STEPS[3].title}
        description={CREATE_WIZARD_STEPS[3].description}
        onSelect={() => {}}
      >
        <ReviewStep
          form={form}
          platform="aws"
          appName="payments"
          branchName="main"
          selection={{
            mode: 'labels',
            group: config.install_groups?.[0] ?? null,
          }}
          inputConfig={reviewInputConfig}
          actions={reviewActions}
          components={reviewComponents}
          submitting={false}
          canSubmit
          onAdvance={() => {}}
          onGoBack={() => {}}
        />
      </CreateInstallWizard>
    </div>
  )
}
