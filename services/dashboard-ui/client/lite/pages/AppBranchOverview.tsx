import { AppBranchOverviewCards } from '../components/organisms/AppBranchOverviewCards'
import { useAppBranchPageChrome } from './AppBranchLayout'

export const AppBranchOverview = () => {
  useAppBranchPageChrome()

  return <AppBranchOverviewCards />
}
