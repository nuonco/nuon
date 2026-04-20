import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import { DevSection } from '../../shared/DevSection'
import { PhoneHomeModal } from '../../PhoneHomeModal'
import { useSurfaces } from '@/hooks/use-surfaces'

interface IDevInstallSection {
  installId?: string
  orgId: string
}

export const DevInstallSection = ({
  installId: routeInstallId,
  orgId,
}: IDevInstallSection) => {
  const { addModal } = useSurfaces()
  const [inputInstallId, setInputInstallId] = useState('')

  const installId = routeInstallId || inputInstallId.trim()

  const handleOpenPhoneHome = () => {
    if (!installId) return
    const modal = (
      <PhoneHomeModal installId={installId} orgId={orgId} />
    )
    addModal(modal)
  }

  return (
    <DevSection
      title="Install controls"
      subtitle={
        routeInstallId ? (
          <div className="flex gap-2">
            Install: <ID>{routeInstallId}</ID>
          </div>
        ) : undefined
      }
    >
      {!routeInstallId && (
        <div className="flex flex-col gap-1">
          <Text variant="label">Install ID</Text>
          <input
            type="text"
            value={inputInstallId}
            onChange={(e) => setInputInstallId(e.target.value)}
            placeholder="Enter install ID"
            className="px-3 py-2 rounded-md border border-gray-200 dark:border-gray-700 bg-transparent text-sm font-mono focus:outline-none focus:border-teal-500"
          />
        </div>
      )}

      <Card className="border-l-4 border-l-teal-500">
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <Icon
              variant="Phone"
              size="20"
              className="text-teal-600 dark:text-teal-400"
            />
            <Text variant="base" weight="strong">
              Phone home
            </Text>
          </div>
          <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-teal-300 dark:hover:border-teal-600 transition-colors">
            <div className="flex flex-col">
              <Text variant="base" weight="strong">
                Send phone home
              </Text>
              <Text
                variant="subtext"
                className="text-gray-600 dark:text-gray-300"
              >
                Send a phone home request with the latest body, or edit it
                before sending
              </Text>
            </div>
            <Button
              onClick={handleOpenPhoneHome}
              variant="secondary"
              disabled={!installId}
            >
              <Icon variant="Phone" />
              Send phone home
            </Button>
          </div>
        </div>
      </Card>
    </DevSection>
  )
}
