'use client'

import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useQueryParams } from '@/hooks/use-query-params'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TFileResponse } from '@/types'
import { downloadFileOnClick } from '@/utils/file-download'

interface IAuditHistory {}

export const AuditHistoryModal = ({ ...props }: IAuditHistory & IModal) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  
  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000),
    end: new Date(),
  })

  const params = useQueryParams({
    start: dateRange.start.toISOString(),
    end: dateRange.end.toISOString(),
  })

  const {
    data: auditLog,
    error,
    isLoading,
  } = useQuery<TFileResponse>({
    dependencies: [params],
    path: `/api/orgs/${org.id}/installs/${install.id}/audit-logs${params}`,
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
        fileType: 'csv',
        mimeType: 'text/csv',
        callback: () => {
          removeModal(props.modalId)
        },
      })
    }
  }


  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
          <Icon variant="ClockClockwise" size="24" />
          Audit History
        </Text>
      }
      primaryActionTrigger={
        isLoading || !auditLog?.content
          ? {
              children: (
                <span className="flex items-center gap-2">
                  <Icon variant="Loading" /> Download CSV
                </span>
              ),
              disabled: true,
              variant: 'primary',
            }
          : {
              children: (
                <span className="flex items-center gap-2">
                  <Icon variant="DownloadSimple" size="18" /> Download CSV
                </span>
              ),
              onClick: handleDownload,
              variant: 'primary',
            }
      }
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              'Unable to load audit logs for the selected date range'}
          </Banner>
        ) : null}
        
        <Text variant="base">
          See a complete record of all activities performed in this install.
        </Text>

        <div className="flex flex-col gap-3">
          <RadioInput
            name="date-range"
            value="1"
            onChange={() => handleDateChange(1)}
            labelProps={{ labelText: "Last one hour" }}
          />
          <RadioInput
            name="date-range"
            value="24"
            onChange={() => handleDateChange(24)}
            labelProps={{ labelText: "Last 24 hours" }}
          />
          <RadioInput
            name="date-range"
            value="168"
            onChange={() => handleDateChange(7 * 24)}
            defaultChecked={true}
            labelProps={{ labelText: "Last week" }}
          />
          <RadioInput
            name="date-range"
            value="720"
            onChange={() => handleDateChange(30 * 24)}
            labelProps={{ labelText: "Last 30 days" }}
          />
          <RadioInput
            name="date-range"
            value="1440"
            onChange={() => handleDateChange(60 * 24)}
            labelProps={{ labelText: "Last 60 days" }}
          />
        </div>
      </div>
    </Modal>
  )
}

export const AuditHistoryButton = ({ ...props }: IAuditHistory & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <AuditHistoryModal />

  return (
    <Button     
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Audit History
      <Icon variant="ClockClockwise" />
    </Button>
  )
}
