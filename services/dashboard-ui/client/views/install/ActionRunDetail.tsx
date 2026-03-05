import { ActionStepGraph } from '@/components/actions/ActionStepsGraph'
import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs'
import { Text } from '@/components/common/Text'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { hydrateActionRunSteps } from '@/utils/action-utils'

export const ActionRunDetail = () => {
  const { installActionRun } = useInstallActionRun()

  const hydratedSteps = hydrateActionRunSteps({
    steps: installActionRun?.steps,
    stepConfigs: installActionRun?.config?.steps,
  })

  return (
    <div className="flex flex-col gap-6">
      {hydratedSteps?.length ? (
        <ActionStepGraph steps={hydratedSteps} />
      ) : null}
      <Text variant="base" weight="strong">Outputs</Text>
      <InstallActionRunOutputs />
    </div>
  )
}
