import { Code } from '@/components/common/Code'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import {
  GroupStep,
  type TGroupStepSelection,
} from '@/components/installs/CreateInstall/GroupStep'
import type { TAppBranchConfig } from '@/types'
import {
  assignmentBlocker,
  enrollmentBlocker,
  labelsMatchGroup,
} from '../readiness'
import { StepActions } from '../StepActions'

interface IAppStep {
  appName: string
  branchName: string
  config?: TAppBranchConfig
  selected: TGroupStepSelection
  labels: { key: string; value: string }[]
  onSelect: (selection: TGroupStepSelection) => void
  onAdvance: () => void
  cancelHref?: string
}

export const AppStep = ({
  appName,
  branchName,
  config,
  selected,
  labels,
  onSelect,
  onAdvance,
  cancelHref,
}: IAppStep) => {
  const groups = config?.install_groups ?? []
  const blocker =
    enrollmentBlocker(groups.length, selected) ??
    assignmentBlocker(selected, labels, config)
  const hasSelectorLabels = groups.some(
    (group) => !group.default && labelsMatchGroup(labels, group)
  )

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-5">
        <div className="grid gap-5 sm:grid-cols-2">
          <LabeledValue label="App">{appName}</LabeledValue>
          <LabeledValue label="Branch">
            <Code variant="inline">{branchName}</Code>
          </LabeledValue>
        </div>
        {config ? (
          <GroupStep
            config={config}
            selected={selected}
            onSelect={onSelect}
            disableExplicit={hasSelectorLabels}
            explicitDisabledReason="These labels already assign the install to a group"
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            This branch has no deployment plan configured. The install will be
            created on the branch without a group assignment.
          </Text>
        )}
      </div>
      <StepActions
        onAdvance={onAdvance}
        label="Continue"
        showCaret
        cancelHref={cancelHref}
        disabled={!!blocker}
        disabledReason={blocker}
      />
    </div>
  )
}
