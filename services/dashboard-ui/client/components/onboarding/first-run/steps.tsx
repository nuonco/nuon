import type { ReactNode } from 'react'
import type { IWizardStepDef } from '@/providers/onboarding-wizard-provider'
import type { TFirstRunStep } from '@/hooks/use-first-run-journey'
import { ConnectStep } from './ConnectStep'
import {
  AWS_QUICK_CREATE_DOCS,
  DOCS_APP_INPUTS,
  DOCS_APPS,
  DOCS_STACKS,
  GCP_INFRA_MANAGER_DOCS,
  KITCHEN_SINK_LABEL,
  type TCloud,
  type TPath,
} from './constants'
import { DeployStep } from './DeployStep'
import { ProvisionStep } from './ProvisionStep'
import { InlineLink } from './shared'
import { StackStep } from './StackStep'
import { StartStep } from './StartStep'

type TFirstRunStepDef = IWizardStepDef & { id: TFirstRunStep }

const START_STEP: TFirstRunStepDef = {
  id: 'start',
  title: 'Create your first app template',
  navLabel: 'Start',
  hideTitle: true,
  component: StartStep,
}

const CONNECT_STEP: TFirstRunStepDef = {
  id: 'connect',
  title: 'Fill in your app template',
  navLabel: 'Connect',
  description: (
    <>
      <InlineLink href={DOCS_APPS} textVariant="body">
        App templates
      </InlineLink>{' '}
      are how Nuon installs and upgrades your app in every customer&apos;s cloud.
    </>
  ),
  component: ConnectStep,
}

const DEPLOY_STEP: TFirstRunStepDef = {
  id: 'deploy',
  title: 'Your app is ready for BYOC',
  navLabel: 'Deploy',
  description: 'Now you can test the flow your customer will see.',
  component: DeployStep,
}

const EXAMPLE_DEPLOY_STEP: TFirstRunStepDef = {
  ...DEPLOY_STEP,
  title: `Test install of ${KITCHEN_SINK_LABEL} demo app`,
  description: (
    <>
      In this example, there is a region picker. We support a{' '}
      <InlineLink href={DOCS_APP_INPUTS} textVariant="body">
        range of configurable inputs
      </InlineLink>
      .
    </>
  ),
}

const STACK_STEP_INTRO: Record<TCloud, ReactNode> = {
  aws: (
    <>
      Nuon is generating a{' '}
      <InlineLink href={AWS_QUICK_CREATE_DOCS} textVariant="body">
        CloudFormation quick-create link
      </InlineLink>
      . This is a common install method for BYOC customers on AWS.
    </>
  ),
  gcp: (
    <>
      Nuon is generating the Terraform for your stack. On Google Cloud, BYOC customers apply it themselves
      or through{' '}
      <InlineLink href={GCP_INFRA_MANAGER_DOCS} textVariant="body">
        Infrastructure Manager
      </InlineLink>
      .
    </>
  ),
  azure: (
    <>
      Nuon is generating the Bicep template and the commands that deploy it. At the default{' '}
      <InlineLink href={DOCS_STACKS} textVariant="body">
        resource group scope
      </InlineLink>
      , BYOC customers create a resource group and Key Vault first, then run the commands.
    </>
  ),
}

const stackStep = (cloud: TCloud): TFirstRunStepDef => ({
  id: 'stack',
  title: 'Create the install stack',
  navLabel: 'Stack',
  description: STACK_STEP_INTRO[cloud],
  component: StackStep,
})

const PROVISION_STEP: TFirstRunStepDef = {
  id: 'provision',
  title: 'Your first BYOC install is deploying',
  navLabel: 'Provision',
  description: 'Follow it live from the CLI, your agent, or the deploy workflow.',
  component: ProvisionStep,
}

export const buildFirstRunSteps = (path: TPath, cloud: TCloud): TFirstRunStepDef[] =>
  path === 'own'
    ? [START_STEP, CONNECT_STEP, DEPLOY_STEP, stackStep(cloud), PROVISION_STEP]
    : [START_STEP, EXAMPLE_DEPLOY_STEP, stackStep(cloud), PROVISION_STEP]

export const stepIndexFor = (steps: TFirstRunStepDef[], step: TFirstRunStep) => {
  const order: TFirstRunStep[] = ['start', 'connect', 'deploy', 'stack', 'provision']
  for (const name of order.slice(order.indexOf(step))) {
    const index = steps.findIndex((s) => s.id === name)
    if (index >= 0) return index
  }
  return steps.length - 1
}
