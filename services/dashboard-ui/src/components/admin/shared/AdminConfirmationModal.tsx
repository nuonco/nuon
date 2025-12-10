'use client'

import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

interface AdminConfirmationModalProps extends IModal {
  title: string
  message: string
  onConfirm: () => void
  variant?: 'default' | 'warning' | 'danger'
  requiresInput?: boolean
  inputText?: string
}

const getWarningIcon = (variant: AdminConfirmationModalProps['variant']) => {
  switch (variant) {
    case 'danger':
      return 'Warning'
    case 'warning':
      return 'Warning'
    default:
      return 'Info'
  }
}

const getBannerTheme = (variant: AdminConfirmationModalProps['variant']) => {
  switch (variant) {
    case 'danger':
      return 'error' as const
    case 'warning':
      return 'warn' as const
    default:
      return 'info' as const
  }
}

export const AdminConfirmationModal = ({
  title,
  message,
  onConfirm,
  variant = 'default',
  requiresInput = false,
  inputText = 'CONFIRM',
  ...props
}: AdminConfirmationModalProps) => {
  const { removeModal } = useSurfaces()
  const [inputValue, setInputValue] = useState('')
  const [isValid, setIsValid] = useState(!requiresInput)

  const handleCancel = () => {
    removeModal(props.modalId)
  }

  const handleConfirm = () => {
    onConfirm()
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading={
        <div className="flex items-center gap-3">
          <Icon variant={getWarningIcon(variant)} />
          <Text variant="h3" weight="strong">{title}</Text>
        </div>
      }
      primaryActionTrigger={{
        children: 'Confirm Action',
        disabled: !isValid,
        onClick: handleConfirm,
        variant: variant === 'danger' ? 'danger' : 'primary'
      }}
      onClose={handleCancel}
      {...props}
    >
      <div className="flex flex-col gap-6">
        <Text variant="body" className="text-gray-700 dark:text-gray-300">
          {message}
        </Text>
        
        {(variant === 'warning' || variant === 'danger') && (
          <Banner theme={getBannerTheme(variant)}>
            {variant === 'danger' 
              ? 'This is a destructive action and cannot be undone.'
              : 'This action will affect running infrastructure.'
            }
          </Banner>
        )}
        
        {requiresInput && (
          <div className="flex flex-col gap-3">
            <Text variant="base" weight="strong">
              Type <span className="text-red-600 dark:text-red-400 font-mono">{inputText}</span> to proceed:
            </Text>
            <Input
              value={inputValue}
              onChange={(e) => {
                setInputValue(e.target.value)
                setIsValid(e.target.value === inputText)
              }}
              placeholder={inputText}
              className="font-mono"
            />
          </div>
        )}
      </div>
    </Modal>
  )
}