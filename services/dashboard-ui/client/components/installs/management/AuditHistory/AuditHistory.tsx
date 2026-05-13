import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IAuditHistoryModal extends IModal {
  error: any
  isLoading: boolean
  hasContent: boolean
  onDownload: () => void
  onDateChange: (hoursAgo: number) => void
}

export const AuditHistoryModal = ({
  error,
  isLoading,
  hasContent,
  onDownload,
  onDateChange,
  ...props
}: IAuditHistoryModal) => {
  return (
    <Modal
      heading={
        <Text
          flex
          className="gap-4"
          variant="h3"
          weight="strong"
        >
          <Icon variant="ClockClockwiseIcon" size="24" />
          Audit history
        </Text>
      }
      primaryActionTrigger={
        isLoading || !hasContent
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
                  <Icon variant="DownloadSimpleIcon" size="18" /> Download CSV
                </span>
              ),
              onClick: onDownload,
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
            onChange={() => onDateChange(1)}
            labelProps={{ labelText: "Last one hour" }}
          />
          <RadioInput
            name="date-range"
            value="24"
            onChange={() => onDateChange(24)}
            labelProps={{ labelText: "Last 24 hours" }}
          />
          <RadioInput
            name="date-range"
            value="168"
            onChange={() => onDateChange(7 * 24)}
            defaultChecked={true}
            labelProps={{ labelText: "Last week" }}
          />
          <RadioInput
            name="date-range"
            value="720"
            onChange={() => onDateChange(30 * 24)}
            labelProps={{ labelText: "Last 30 days" }}
          />
          <RadioInput
            name="date-range"
            value="1440"
            onChange={() => onDateChange(60 * 24)}
            labelProps={{ labelText: "Last 60 days" }}
          />
        </div>
      </div>
    </Modal>
  )
}
