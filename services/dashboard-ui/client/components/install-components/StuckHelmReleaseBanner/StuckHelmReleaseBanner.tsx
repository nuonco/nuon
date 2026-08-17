import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { RecoverHelmReleaseButton } from '@/components/install-components/management/RecoverHelmRelease'
import type { TComponent } from '@/types'

interface IStuckHelmReleaseBanner {
  component: TComponent
  status?: string
}

export const StuckHelmReleaseBanner = ({ component, status }: IStuckHelmReleaseBanner) => {
  return (
    <Banner theme="warn">
      <div className="flex items-center gap-8">
        <div className="flex flex-col max-w-sm">
          <Text weight="strong" variant="base">
            Helm release is stuck
          </Text>
          <Text className="text-pretty" theme="neutral">
            The last deploy could not run because Helm left this release
            {status ? ` in ${status}` : ''} part-way through an earlier operation. Deploys will
            keep failing until the release is recovered.
          </Text>
        </div>
        <RecoverHelmReleaseButton
          className="ml-auto"
          component={component}
          status={status}
          variant="primary"
        />
      </div>
    </Banner>
  )
}
