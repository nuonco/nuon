import { useContext } from 'react'
import { ConfigContext } from '@/providers/config-provider'

export function useConfig() {
  const ctx = useContext(ConfigContext)
  if (!ctx) {
    throw new Error('useConfig must be used within a ConfigProvider')
  }
  return ctx
}
