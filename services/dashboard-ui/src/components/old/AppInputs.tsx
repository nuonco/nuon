'use client'

import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  ArrowsOutSimpleIcon,
  CheckIcon,
  MinusIcon,
} from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Expand } from '@/components/old/Expand'
import { Modal } from '@/components/old/Modal'
import { ToolTip } from '@/components/old/ToolTip'
import { Text, Truncate } from '@/components/old/Typography'
import type { TAppInputConfig } from '@/types'

export interface IAppInputConfig {
  inputConfig: TAppInputConfig
  isNotTruncated?: boolean
}

export const AppInputConfig: FC<IAppInputConfig> = ({
  inputConfig,
  isNotTruncated = false,
}) => {
  return inputConfig && inputConfig?.input_groups?.length ? (
    <div className="flex flex-col gap-2">
      <div className="grid grid-cols-8 gap-4 px-3 py-2 text-cool-grey-600 dark:text-cool-grey-500 text-base">
        <Text className="!font-medium col-span-2">Name</Text>
        <Text className="!font-medium col-span-2">Description</Text>
        <Text className="!font-medium col-span-2">Default</Text>
        <Text className="!font-medium">Required</Text>
        <Text className="!font-medium">Sensitive</Text>
      </div>
      {inputConfig.input_groups.map((group, i) => (
        <Expand
          className="w-full"
          key={`${group.id}-${i}`}
          hasHeadingStyle
          heading={
            <div className="px-3 py-2 text-base grid grid-cols-8 gap-4 items-start">
              <Text className="col-span-2 !font-medium">
                {group.display_name}
              </Text>
              <Text className="col-span-6 text-sm">{group.description}</Text>
            </div>
          }
          id={`${group.id}-${i}`}
          isOpen
          expandContent={
            <div className="divide-y">
              {group.app_inputs.map((input, ii) => (
                <div
                  key={`${input.id}-${ii}`}
                  className="grid grid-cols-8 gap-4 px-3 py-4 items-start"
                >
                  <div className="col-span-2 gap-2 flex flex-col items-start justify-start">
                    <Text className="!font-medium">{input.display_name}</Text>{' '}
                    {input?.name?.length >= 14 && !isNotTruncated ? (
                      <ToolTip tipContent={input.name}>
                        <Text className="font-mono text-cool-grey-600 dark:text-cool-grey-500 text-sm">
                          <Truncate variant="small">{input.name}</Truncate>
                        </Text>
                      </ToolTip>
                    ) : (
                      <Text className="font-mono text-cool-grey-600 dark:text-cool-grey-500 text-sm">
                        {input.name}
                      </Text>
                    )}
                  </div>

                  <Text className="col-span-2 text-sm">
                    {input.description}
                  </Text>

                  <div className="col-span-2 gap-2 flex flex-col items-start justify-start">
                    {input?.default?.length >= 14 && !isNotTruncated ? (
                      <ToolTip tipContent={input.default}>
                        <Text className="text-sm">
                          <Truncate variant="small">{input.default}</Truncate>
                        </Text>
                      </ToolTip>
                    ) : (
                      <Text className="text-sm">
                        {input.default || <MinusIcon />}
                      </Text>
                    )}
                  </div>

                  <Text className="text-sm">
                    {input.required ? <CheckIcon /> : <MinusIcon />}
                  </Text>

                  <Text className="text-sm">
                    {input.sensitive ? <CheckIcon /> : <MinusIcon />}
                  </Text>
                </div>
              ))}
            </div>
          }
        />
      ))}
    </div>
  ) : (
    <Text>No app inputs configured</Text>
  )
}

export const AppInputConfigModal: FC<IAppInputConfig & { appName: string }> = ({
  appName,
  inputConfig,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <>
      {isOpen
        ? createPortal(
            <Modal
              heading={`${appName} inputs`}
              isOpen={isOpen}
              onClose={() => {
                setIsOpen(false)
              }}
            >
              <AppInputConfig inputConfig={inputConfig} isNotTruncated />
            </Modal>,
            document.body
          )
        : null}
      <Button
        className="text-sm !font-medium flex items-center gap-2 !p-1"
        onClick={() => {
          setIsOpen(true)
        }}
        title="Expand install inputs"
        variant="ghost"
      >
        <ArrowsOutSimpleIcon />
      </Button>
    </>
  )
}
