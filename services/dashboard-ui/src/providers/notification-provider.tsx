'use client'

import { createContext, useCallback, useEffect, useState, type ReactNode } from 'react'

interface NotificationOptions {
  title: string
  body?: string
  icon?: string
  tag?: string
  requireInteraction?: boolean
  silent?: boolean
  data?: any
  sound?: string
  onClick?: (data?: any) => void
}

interface NotificationSettings {
  permissionRequested: boolean
  permissionGrantedAt?: string
  lastPermissionRequest?: string
}

interface NotificationContextType {
  emitNotification: (options: NotificationOptions) => Promise<boolean>
  permission: NotificationPermission
  requestPermission: () => Promise<NotificationPermission>
  isSupported: boolean
  settings: NotificationSettings
  hasRequestedPermission: boolean
}

export const NotificationContext = createContext<NotificationContextType | null>(null)

const STORAGE_KEY = 'notification_settings'

export function NotificationProvider({ 
  children,
  autoRequestOnLoad = true,
  autoRequestDelay = 2000
}: { 
  children: ReactNode
  autoRequestOnLoad?: boolean
  autoRequestDelay?: number
}) {
  const [permission, setPermission] = useState<NotificationPermission>('default')
  const [isSupported, setIsSupported] = useState(false)
  const [settings, setSettings] = useState<NotificationSettings>({
    permissionRequested: false
  })
  const [hasRequestedPermission, setHasRequestedPermission] = useState(false)

  // Load settings from localStorage
  useEffect(() => {
    if (typeof window === 'undefined') return

    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsedSettings = JSON.parse(stored)
        setSettings(parsedSettings)
        setHasRequestedPermission(parsedSettings.permissionRequested)
      }
    } catch (error) {
      console.warn('Failed to load notification settings from localStorage:', error)
    }
  }, [])

  // Check browser support and current permission
  useEffect(() => {
    if (typeof window === 'undefined') return

    if ('Notification' in window) {
      setIsSupported(true)
      setPermission(Notification.permission)
    }
  }, [])

  // Auto-request permission on page load
  useEffect(() => {
    if (!isSupported || !autoRequestOnLoad) return
    if (permission !== 'default') return
    if (hasRequestedPermission) return

    const timer = setTimeout(async () => {
      await requestPermission()
    }, autoRequestDelay)

    return () => clearTimeout(timer)
  }, [isSupported, permission, hasRequestedPermission, autoRequestOnLoad, autoRequestDelay])

  // Save settings to localStorage
  const saveSettings = useCallback((newSettings: NotificationSettings) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newSettings))
      setSettings(newSettings)
    } catch (error) {
      console.warn('Failed to save notification settings to localStorage:', error)
    }
  }, [])

  const requestPermission = useCallback(async (): Promise<NotificationPermission> => {
    if (!isSupported) {
      console.warn('Notifications are not supported in this browser')
      return 'denied'
    }

    try {
      const result = await Notification.requestPermission()
      setPermission(result)
      setHasRequestedPermission(true)

      const now = new Date().toISOString()
      const newSettings: NotificationSettings = {
        permissionRequested: true,
        lastPermissionRequest: now,
        ...(result === 'granted' && { permissionGrantedAt: now })
      }

      saveSettings(newSettings)
      
      return result
    } catch (error) {
      console.error('Error requesting notification permission:', error)
      return 'denied'
    }
  }, [isSupported, saveSettings])

  const emitNotification = useCallback(async (options: NotificationOptions): Promise<boolean> => {
    if (!isSupported) {
      console.warn('Notifications are not supported')
      return false
    }

    if (permission !== 'granted') {
      console.warn('Notification permission not granted')
      return false
    }

    try {
      const notification = new Notification(options.title, {
        body: options.body,
        icon: options.icon,
        tag: options.tag,
        requireInteraction: options.requireInteraction,
        silent: options.silent,
        data: options.data,
      })

      // Handle click on the notification
      if (options.onClick) {
        notification.onclick = () => {
          options.onClick!(options.data)
          notification.close()
        }
      }

      // Play sound if provided
      if (options.sound) {
        try {
          const audio = new Audio(options.sound)
          await audio.play()
        } catch (soundError) {
          console.warn('Failed to play notification sound:', soundError)
        }
      }

      // Auto close after 5 seconds unless requireInteraction is true
      if (!options.requireInteraction) {
        setTimeout(() => {
          notification.close()
        }, 10000)
      }

      return true
    } catch (error) {
      console.error('Error creating notification:', error)
      return false
    }
  }, [isSupported, permission])

  const value: NotificationContextType = {
    emitNotification,
    permission,
    requestPermission,
    isSupported,
    settings,
    hasRequestedPermission,
  }

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  )
}
