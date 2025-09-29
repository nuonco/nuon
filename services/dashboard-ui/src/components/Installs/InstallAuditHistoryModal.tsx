'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { DownloadSimpleIcon, ClockClockwiseIcon } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { RadioInput } from '@/components/Input'
import { Modal } from '@/components/Modal'
import { Text } from '@/components/Typography'

interface IInstallAuditHistoryModal {
  installId: string
  orgId: string
}

// TODO(nnnnat): refactor to new api/orgs endpoint
export const InstallAuditHistoryModal: FC<IInstallAuditHistoryModal> = ({
  installId,
  orgId,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [dateRange, setDateRange] = useState({
    start: new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000), // 7 days ago
    end: new Date(),
  })

  const handleDateChange = (hoursAgo: number) => {
    const end = new Date()
    const start = new Date(end.getTime() - (hoursAgo * 60 * 60 * 1000))
    setDateRange({ start, end })
  }

  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              className="max-w-lg"
              heading="Audit History"
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <div className="flex flex-col gap-3 mb-6">
                <Text variant="reg-14" className="leading-relaxed">
                  See a complete record of all activities performed in this install.
                </Text>
                <RadioInput
                  className="mt-0.5"
                  key={"last-1-hour"}
                  name="date-range"
                  value={"1"}
                  onChange={() => {
                    handleDateChange(1)
                  }}
                  labelClassName="!px-6 !items-start"
                  labelText={
                    <span className="flex flex-col gap-0">
                      <Text variant="med-12">Last one hour</Text>
                    </span>
                  }
                />
                <RadioInput
                  className="mt-0.5"
                  key={"last-24-hours"}
                  name="date-range"
                  value={"24"}
                  onChange={() => {
                    handleDateChange(24)
                  }}
                  labelClassName="!px-6 !items-start"
                  labelText={
                    <span className="flex flex-col gap-0">
                      <Text variant="med-12">Last 24 hours</Text>
                    </span>
                  }
                />
                <RadioInput
                  className="mt-0.5"
                  key={"last-week"}
                  name="date-range"
                  value={"168"}
                  onChange={() => {
                    handleDateChange(7*24) // 168, 7 days in hours
                  }}
                  defaultChecked={true}
                  labelClassName="!px-6 !items-start"
                  labelText={
                    <span className="flex flex-col gap-0">
                      <Text variant="med-12">Last week</Text>
                    </span>
                  }
                />
                <RadioInput
                  className="mt-0.5"
                  key={"last-30-days"}
                  name="date-range"
                  value={"720"}
                  onChange={() => {
                    handleDateChange(30*24) // 720, 30 days in hours
                  }}
                  labelClassName="!px-6 !items-start"
                  labelText={
                    <span className="flex flex-col gap-0">
                      <Text variant="med-12">Last 30 days</Text>
                    </span>
                  }
                />
                <RadioInput
                  className="mt-0.5"
                  key={"last-60-days"}
                  name="date-range"
                  value={"1440"}
                  onChange={() => {
                    handleDateChange(60*24) // 1440, 60 days in hours
                  }}
                  labelClassName="!px-6 !items-start"
                  labelText={
                    <span className="flex flex-col gap-0">
                      <Text variant="med-12">Last 60 days</Text>
                    </span>
                  }
                />
              </div>
              <div className="flex gap-3 justify-end">
                <Button
                  onClick={() => {
                    setIsOpen(false)
                  }}
                  className="text-sm"
                >
                  Cancel
                </Button>
                <a href={
                  `/api/${orgId}/installs/${installId}/audit-logs?`+
                  `start=${dateRange.start.toISOString()}&end=${dateRange.end.toISOString()}`
                } target="_blank" rel="noopener noreferrer">
                  <Button
                    className="text-sm flex items-center gap-1"
                    variant="primary"
                    onClick={() => { setIsOpen(false) }}
                  >
                    <DownloadSimpleIcon size="18" />
                    {' '}
                    Download CSV
                  </Button>
                </a>
              </div>
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full"
        variant="ghost"
        onClick={() => {
          setIsOpen(true)
        }}
      >
        <ClockClockwiseIcon size="16" />
        Audit History
      </Button>
    </>
  )
}
