import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IAdminConfirmationModal extends IModal {
  title: string
  message: string
  onConfirm: () => void
  onCancel: () => void
  variant?: 'default' | 'warning' | 'danger'
  requiresInput?: boolean
  inputText?: string
  isPending?: boolean
}

const getWarningIcon = (variant: IAdminConfirmationModal['variant']) => {
  switch (variant) {
    case 'danger':
      return 'WarningIcon'
    case 'warning':
      return 'WarningIcon'
    default:
      return 'InfoIcon'
  }
}

const getBannerTheme = (variant: IAdminConfirmationModal['variant']) => {
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
  onCancel,
  variant = 'default',
  requiresInput = false,
  inputText = 'CONFIRM',
  isPending = false,
  ...props
}: IAdminConfirmationModal) => {
  const [inputValue, setInputValue] = useState('')
  const [isValid, setIsValid] = useState(!requiresInput)

  return (
    <Modal
      heading={
        <div className="flex items-center gap-3">
          <Icon variant={getWarningIcon(variant)} />
          <Text variant="h3" weight="strong">{title}</Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Executing...
          </span>
        ) : 'Confirm action',
        disabled: !isValid || isPending,
        onClick: onConfirm,
        variant: variant === 'danger' ? 'danger' : 'primary'
      }}
      onClose={onCancel}
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
