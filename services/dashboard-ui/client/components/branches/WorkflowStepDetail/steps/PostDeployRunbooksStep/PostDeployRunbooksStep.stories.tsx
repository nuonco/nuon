export default {
  title: 'Branches/WorkflowStepDetail/PostDeployRunbooksStep',
}

import type { ComponentProps } from 'react'
import { StepCardStory } from '@/components/__stories__/helpers'
import { PostDeployRunbooksStep } from './PostDeployRunbooksStep'
import {
  InstallRunbooksRow,
  type IInstallRunbooksRow,
} from './InstallRunbooksRow'

const StepInCard = (props: ComponentProps<typeof PostDeployRunbooksStep>) => (
  <StepCardStory name="post-deploy runbooks">
    <PostDeployRunbooksStep {...props} />
  </StepCardStory>
)

const mkInstall = (over: Record<string, any> = {}): any => ({
  id: 'ins_acme',
  name: 'httpbin1',
  cloud_platform: 'aws',
  aws_account: { region: 'us-west-2' },
  ...over,
})

const wfHref = '/org_1/installs/ins_acme/workflows/wf_1'
const installHref = '/org_1/installs/ins_acme'

const okRunbooks = [
  {
    runbookId: 'rb1',
    runbookName: 'verify_status',
    status: 'success',
    workflowHref: wfHref,
  },
  {
    runbookId: 'rb2',
    runbookName: 'deploy_verify_and_logs',
    status: 'success',
    workflowHref: wfHref,
  },
]

export const RowAllSucceeded = () => (
  <InstallRunbooksRow
    installId="ins_acme"
    install={mkInstall()}
    installHref={installHref}
    runbooks={okRunbooks}
  />
)

export const RowFailedSecondRunbook = () => (
  <InstallRunbooksRow
    installId="ins_acme"
    install={mkInstall()}
    installHref={installHref}
    runbooks={[
      okRunbooks[0],
      { ...okRunbooks[1], runbookName: 'smoke_test', status: 'error' },
    ]}
  />
)

export const RowInProgress = () => (
  <InstallRunbooksRow
    installId="ins_acme"
    install={mkInstall()}
    installHref={installHref}
    runbooks={[okRunbooks[0], { ...okRunbooks[1], status: 'in-progress' }]}
  />
)

const rows: IInstallRunbooksRow[] = [
  {
    installId: 'ins_acme',
    install: mkInstall(),
    installHref,
    runbooks: okRunbooks,
  },
  {
    installId: 'ins_globex',
    install: mkInstall({ id: 'ins_globex', name: 'httpbin2' }),
    installHref,
    runbooks: okRunbooks,
  },
]

export const StepAllSucceeded = () => (
  <StepInCard
    groupName="prod"
    runbookNames={['verify_status', 'deploy_verify_and_logs']}
    rows={rows}
  />
)

export const StepSingleInstall = () => (
  <StepInCard
    groupName="prod"
    runbookNames={['verify_status', 'deploy_verify_and_logs']}
    rows={[rows[0]]}
  />
)

export const StepWithFailure = () => (
  <StepInCard
    groupName="prod"
    runbookNames={['verify_status', 'smoke_test']}
    rows={[
      rows[0],
      {
        ...rows[1],
        runbooks: [
          okRunbooks[0],
          { ...okRunbooks[1], runbookName: 'smoke_test', status: 'error' },
        ],
      },
    ]}
  />
)

export const StepStarting = () => (
  <StepInCard
    groupName="prod"
    runbookNames={[]}
    rows={[]}
    emptyMessage="Starting post-deploy runbooks"
  />
)
