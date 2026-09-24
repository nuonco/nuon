import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { adminGetOrgFeaturesList, adminUpdateOrgFeatures } from '@/lib'
import { useCurrentUser } from '../../../hooks/use-current-user'
import { useToast } from '../../../hooks/use-toast'
import { useModules } from '../../../providers/modules-provider'
import { useOrg } from '../../../providers/org-provider'
import {
  MODULES,
  MODULE_IDS,
  featuresPatch,
  moduleById,
  moduleFeature,
  presetById,
  presetModules,
  withModule,
  type TModuleId,
  type TModulePresetId,
} from '../../../utils/modules'
import {
  ModuleManager,
  type IModuleState,
  type TModulePending,
} from './ModuleManager'

export const ModuleManagerContainer = () => {
  const { org, orgId, loading } = useOrg()
  const { enabled } = useModules()
  const { user } = useCurrentUser()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [pending, setPending] = useState<TModulePending>()

  const { data: flags } = useQuery({
    queryKey: ['admin-org-features'],
    queryFn: adminGetOrgFeaturesList,
    staleTime: 60_000,
  })

  const mutation = useMutation({
    mutationFn: (features: Record<string, boolean>) =>
      adminUpdateOrgFeatures({
        orgId: orgId!,
        features,
        adminEmail: user.email ?? '',
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['org', orgId] })
    },
    onSettled: () => setPending(undefined),
  })

  const orgName = org?.name ?? 'this org'

  const apply = (
    next: ReadonlySet<TModuleId>,
    key: TModulePending,
    onDone: () => void
  ) => {
    const patch = featuresPatch(enabled, next)
    if (!Object.keys(patch).length || !orgId) return
    setPending(key)
    mutation.mutate(patch, { onSuccess: onDone })
  }

  const toggle = (id: TModuleId, on: boolean) =>
    apply(withModule(enabled, id, on), id, () => {
      const name = moduleById(id)?.name ?? id
      addToast({
        heading: on ? 'Module enabled' : 'Module hidden',
        description: on
          ? `${name} is back in the dashboard for everyone in ${orgName}.`
          : `${name} is hidden from the dashboard for everyone in ${orgName}.`,
        theme: 'success',
      })
    })

  const applyPreset = (id: TModulePresetId) => {
    const next = presetModules(id)
    apply(next, 'preset', () => {
      addToast({
        heading: 'Preset applied',
        description: `${presetById(id)?.name ?? id} enables ${next.size} of ${MODULE_IDS.length} modules for ${orgName}.`,
        theme: 'success',
      })
    })
  }

  const flagByName = new Map(flags?.map((flag) => [flag.name, flag]))
  const states: IModuleState[] = MODULES.map((module) => {
    const flag = flagByName.get(moduleFeature(module.id))
    return {
      module,
      enabled: enabled.has(module.id),
      pinned: !!flag?.forced,
      registered: !flags || !!flag,
    }
  })

  return (
    <ModuleManager
      modules={states}
      onToggle={toggle}
      onPreset={applyPreset}
      pending={pending}
      loading={loading}
      error={mutation.error}
    />
  )
}
