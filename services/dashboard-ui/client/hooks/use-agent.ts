import { useContext } from 'react'
import { AgentContext, type TAgentContext } from '@/providers/agent-provider'

export function useAgent(): TAgentContext {
  const ctx = useContext(AgentContext)
  if (!ctx) {
    throw new Error('useAgent must be used within AgentProvider')
  }
  return ctx
}
