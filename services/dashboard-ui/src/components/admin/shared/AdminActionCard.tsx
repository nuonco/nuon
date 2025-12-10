'use client'

import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Card } from '@/components/common/Card'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminConfirmationModal } from './AdminConfirmationModal'

export interface AdminActionCardProps {
  title: string
  description: string
  action: () => Promise<any>
  variant?: 'default' | 'warning' | 'danger'
  requiresConfirmation?: boolean
  confirmationText?: string
  requiresInput?: boolean
  inputText?: string
}

const getActionIcon = (title: string, variant: AdminActionCardProps['variant']) => {
  // Check for specific action types first
  if (title.toLowerCase().includes('add') || title.toLowerCase().includes('support user')) return 'UserPlus'
  if (title.toLowerCase().includes('remove') || title.toLowerCase().includes('support user')) return 'UserMinus'
  if (title.toLowerCase().includes('reprovision')) return 'ArrowClockwise'
  if (title.toLowerCase().includes('restart')) return 'ArrowCounterClockwise'
  if (title.toLowerCase().includes('teardown') || title.toLowerCase().includes('force')) return 'Trash'
  if (title.toLowerCase().includes('shutdown') && title.toLowerCase().includes('graceful')) return 'Power'
  if (title.toLowerCase().includes('shutdown') || title.toLowerCase().includes('stop')) return 'Stop'
  if (title.toLowerCase().includes('invalidate') || title.toLowerCase().includes('token')) return 'Key'
  if (title.toLowerCase().includes('debug')) return 'Bug'
  if (title.toLowerCase().includes('update') || title.toLowerCase().includes('sandbox')) return 'Upload'
  
  // Fallback to variant-based icons
  switch (variant) {
    case 'warning':
      return 'Warning'
    case 'danger':
      return 'Warning'
    default:
      return 'Play'
  }
}

const getVariantStyles = (variant: AdminActionCardProps['variant']) => {
  switch (variant) {
    case 'warning':
      return {
        buttonVariant: 'secondary' as const
      }
    case 'danger':
      return {
        buttonVariant: 'danger' as const
      }
    default:
      return {
        buttonVariant: 'secondary' as const
      }
  }
}

export const AdminActionCard = ({
  title,
  description,
  action,
  variant = 'default',
  requiresConfirmation = false,
  confirmationText,
  requiresInput = false,
  inputText = 'CONFIRM'
}: AdminActionCardProps) => {
  const { addModal } = useSurfaces()
  const styles = getVariantStyles(variant)

  const { execute, isLoading, data, error } = useServerAction({ action })

  useServerActionToast({
    data,
    error,
    successContent: <Text>{title} completed successfully</Text>,
    successHeading: 'Action Complete',
    errorContent: <Text>Failed to {title.toLowerCase()}. Please try again.</Text>,
    errorHeading: 'Action Failed'
  })

  const handleClick = () => {
    if (requiresConfirmation) {
      const modalId = `admin-confirm-${title.toLowerCase().replace(/\s+/g, '-')}-${Date.now()}`
      const confirmationModal = (
        <AdminConfirmationModal
          modalId={modalId}
          title={`Confirm: ${title}`}
          message={confirmationText || `Are you sure you want to ${title.toLowerCase()}?`}
          onConfirm={() => {
            execute()
          }}
          variant={variant}
          requiresInput={requiresInput}
          inputText={inputText}
        />
      )
      addModal(confirmationModal)
    } else {
      execute()
    }
  }

  return (
    <div className="space-y-3 p-4 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      <div className="flex flex-col">
        <Text variant="base" weight="strong">
          {title}
        </Text>
        <Text variant="subtext" className="text-gray-600 dark:text-gray-300">
          {description}
        </Text>
      </div>
      
      <Button
        onClick={handleClick}
        disabled={isLoading}
        variant={styles.buttonVariant}
      >
        {isLoading ? (
          <>
            <Icon variant="Loading" className="animate-spin" />
            Executing...
          </>
        ) : (
          <>
            <Icon variant={getActionIcon(title, variant)} />
            {title}
          </>
        )}
      </Button>
    </div>
  )
}