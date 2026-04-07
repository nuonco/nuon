import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallAuditLog } from '@/lib'
import { downloadFileOnClick } from '@/utils/file-download'
import { slugify } from '@/utils/string-utils'
import { AuditHistoryModal } from './AuditHistory'

export const AuditHistoryModalContainer = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()

  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000),
    end: new Date(),
  })

  const {
    data: auditLog,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['install-audit-log', org.id, install.id, dateRange.start.toISOString(), dateRange.end.toISOString()],
    queryFn: () =>
      getInstallAuditLog({
        orgId: org.id,
        installId: install.id,
        start: dateRange.start.toISOString(),
        end: dateRange.end.toISOString(),
      }),
  })

  const handleDateChange = (hoursAgo: number) => {
    const end = new Date()
    const start = new Date(end.getTime() - hoursAgo * 60 * 60 * 1000)
    setDateRange({ start, end })
  }

  const handleDownload = () => {
    if (auditLog?.content) {
      downloadFileOnClick({
        ...auditLog,
        filename: `${slugify(install.name)}-audit-log.csv`,
        fileType: 'csv',
        mimeType: 'text/csv',
        callback: () => {
          removeModal(props.modalId)
        },
      })
    }
  }

  return (
    <AuditHistoryModal
      error={error}
      isLoading={isLoading}
      hasContent={!!auditLog?.content}
      onDownload={handleDownload}
      onDateChange={handleDateChange}
      {...props}
    />
  )
}

export const AuditHistoryButton = ({ ...props }: IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <AuditHistoryModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Audit history
      <Icon variant="ClockClockwise" />
    </Button>
  )
}
