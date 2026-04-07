export default {
  title: 'Layout/LandingLayout',
}

import { LandingLayout } from './LandingLayout'
import { Text } from '@/components/common/Text'

export const Default = () => (
  <LandingLayout>
    <div className="flex flex-col gap-6">
      <Text variant="h1" weight="stronger">Welcome to Nuon</Text>
      <Text variant="body" theme="neutral">Deploy your applications to any cloud.</Text>
    </div>
  </LandingLayout>
)
