import { InstallOverviewCards } from '../components/organisms/InstallOverviewCards'
import { useInstallPageChrome } from './InstallLayout'

export const InstallOverview = () => {
  useInstallPageChrome()

  return <InstallOverviewCards />
}
