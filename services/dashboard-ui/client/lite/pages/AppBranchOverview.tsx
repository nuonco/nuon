import { AppBranchOverviewCards } from '../components/organisms/AppBranchOverviewCards'
import { DeploymentPlanStages } from '../components/organisms/DeploymentPlanStages'
import { useAppBranchPageChrome } from './AppBranchLayout'

export const AppBranchOverview = () => {
  useAppBranchPageChrome()

  return (
    <div className="flex flex-col gap-6">
      <AppBranchOverviewCards />
      <DeploymentPlanStages />
    </div>
  )
}
