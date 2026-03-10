import { ActionStepGraph } from '@/components/actions/ActionStepsGraph'
import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { PageTitle } from '@/components/navigation/PageTitle'
import { hydrateActionRunSteps } from '@/utils/action-utils'

export const ActionRunDetail = () => {
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()

  const hydratedSteps = hydrateActionRunSteps({
    steps: installActionRun?.steps,
    stepConfigs: installActionRun?.config?.steps,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageTitle title={`${installActionRun?.trigger_type ? `${installActionRun.trigger_type} run` : 'Run'} | ${install?.name}`} />
      {hydratedSteps?.length ? (
        <ActionStepGraph steps={hydratedSteps} />
      ) : null}
      <Text variant="base" weight="strong">Outputs</Text>
      <InstallActionRunOutputs />
    </div>
  )
}
