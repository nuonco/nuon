import React, { type FC } from 'react'
import { FaDocker } from 'react-icons/fa'
import { GoQuestion } from 'react-icons/go'
import {
  SiAwslambda,
  SiHelm,
  SiKubernetes,
  SiOpencontainersinitiative,
  SiTerraform,
} from 'react-icons/si'

export type TComponentConfigType =
  | 'docker_build'
  | 'external_image'
  | 'terraform_module'
  | 'helm_chart'
  | 'job'
  | 'kubernetes_manifest'
  | 'unknown'

export const ComponentConfigType: FC<{
  configType: TComponentConfigType
  isIconOnly?: boolean
}> = ({ configType, isIconOnly = false }) => {
  let cfgType = {}
  switch (configType) {
    case 'docker_build':
      cfgType = { icon: <FaDocker />, name: 'Docker' }
      break
    case 'external_image':
      cfgType = { icon: <SiOpencontainersinitiative />, name: 'External image' }
      break
    case 'helm_chart':
      cfgType = { icon: <SiHelm />, name: 'Helm' }
      break
    case 'terraform_module':
      cfgType = { icon: <SiTerraform />, name: 'Terraform' }
      break
    case 'job':
      cfgType = { icon: <SiAwslambda />, name: 'Job' }
      break
    case 'kubernetes_manifest':
      cfgType = { icon: <SiKubernetes />, name: 'Kubernetes Manifest' }
      break
    default:
      cfgType = { icon: <GoQuestion />, name: 'Unknown' }
  }

  return (
    <span className="flex items-center gap-1 text-nowrap">
      {cfgType['icon']} {!isIconOnly && cfgType['name']}
    </span>
  )
}
