import { matchesSelector } from '@/components/match/matches'
import type { TAppBranchInstallGroup, TInstall } from '@/types'
import { installAppBranchGroup } from './active-app-branch-connection'

type InstallGroup = Pick<
  TAppBranchInstallGroup,
  'name' | 'default' | 'label_selector'
>

export interface InstallGroupMembership {
  installsByGroup: TInstall[][]
  unassignedInstalls: TInstall[]
  overlappingInstalls: TInstall[]
}

export const resolveInstallGroupMembership = (
  installs: TInstall[],
  groups: InstallGroup[]
): InstallGroupMembership => {
  const installsByGroup = groups.map(() => [] as TInstall[])
  const unassignedInstalls: TInstall[] = []
  const overlappingInstalls: TInstall[] = []
  const defaultGroupIndex = groups.findIndex((group) => group.default)

  installs.forEach((install) => {
    const explicitGroup = installAppBranchGroup(install)
    if (explicitGroup) {
      const explicitGroupIndex = groups.findIndex(
        (group) => group.name === explicitGroup
      )
      if (explicitGroupIndex === -1) {
        unassignedInstalls.push(install)
      } else {
        installsByGroup[explicitGroupIndex].push(install)
      }
      return
    }

    const matchingGroupIndexes = groups.flatMap((group, index) =>
      Object.keys(group.label_selector?.match_labels ?? {}).length > 0 &&
      matchesSelector(install.labels, group.label_selector)
        ? [index]
        : []
    )

    if (matchingGroupIndexes.length > 0) {
      matchingGroupIndexes.forEach((index) =>
        installsByGroup[index].push(install)
      )
      if (matchingGroupIndexes.length > 1) {
        overlappingInstalls.push(install)
      }
      return
    }

    if (defaultGroupIndex === -1) {
      unassignedInstalls.push(install)
    } else {
      installsByGroup[defaultGroupIndex].push(install)
    }
  })

  return { installsByGroup, unassignedInstalls, overlappingInstalls }
}
