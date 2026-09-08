import { Card } from '../components/atoms/Card'
import { Text } from '../components/atoms/Text'
import { useInstallPageChrome } from './InstallLayout'

export const InstallActivity = () => {
  useInstallPageChrome('Activity')

  return (
    <Card className="min-h-40">
      <Text variant="caption" color="tertiary">
        Page content will be added in a follow-up.
      </Text>
    </Card>
  )
}
