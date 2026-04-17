import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/use-auth'
import {
  adminListSandboxRunners,
  adminGetSandboxConfigs,
  adminGetSandboxJobs,
  adminGetSandboxTemplates,
} from '@/lib'
import { AdminSandboxSection } from './AdminSandboxSection'

const EMPTY_TEMPLATES = { log_templates: [], plan_templates: [] }

export const AdminSandboxSectionContainer = () => {
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const queryClient = useQueryClient()
  const [selectedRunnerId, setSelectedRunnerId] = useState<string | null>(null)

  const { data: runners = [], isLoading: runnersLoading } = useQuery({
    queryKey: ['sandbox-runners'],
    queryFn: adminListSandboxRunners,
    refetchInterval: 10_000,
  })

  const { data: templates = EMPTY_TEMPLATES } = useQuery({
    queryKey: ['sandbox-templates'],
    queryFn: adminGetSandboxTemplates,
    staleTime: Infinity,
  })

  const { data: configs = [], isLoading: configsLoading } = useQuery({
    queryKey: ['sandbox-configs', selectedRunnerId],
    queryFn: () => adminGetSandboxConfigs({ runnerId: selectedRunnerId! }),
    enabled: !!selectedRunnerId,
  })

  const { data: jobs = [], isLoading: jobsLoading } = useQuery({
    queryKey: ['sandbox-jobs', selectedRunnerId],
    queryFn: () => adminGetSandboxJobs({ runnerId: selectedRunnerId! }),
    enabled: !!selectedRunnerId,
    refetchInterval: 5_000,
  })

  const handleRefreshConfigs = () => {
    queryClient.invalidateQueries({ queryKey: ['sandbox-configs', selectedRunnerId] })
  }

  return (
    <AdminSandboxSection
      runners={runners}
      runnersLoading={runnersLoading}
      selectedRunnerId={selectedRunnerId}
      onSelectRunner={setSelectedRunnerId}
      configs={configs}
      configsLoading={configsLoading}
      jobs={jobs}
      jobsLoading={jobsLoading}
      templates={templates}
      adminEmail={adminEmail}
      onRefreshConfigs={handleRefreshConfigs}
    />
  )
}
