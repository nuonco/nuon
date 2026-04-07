import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'
import { BuildHeader } from './BuildHeader'
import type { TComponent } from '@/types'

export const BuildHeaderContainer = ({ component }: { component: TComponent }) => {
  const { app } = useApp()
  const { build } = useBuild()

  return <BuildHeader component={component} build={build} app={app} />
}
