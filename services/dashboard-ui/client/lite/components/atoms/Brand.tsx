import { AWSBrand } from './brands/AWSBrand'
import { AzureBrand } from './brands/AzureBrand'
import { DockerBrand } from './brands/DockerBrand'
import { GCPBrand } from './brands/GCPBrand'
import { GitHubBrand } from './brands/GitHubBrand'
import { HelmBrand } from './brands/HelmBrand'
import { KubernetesBrand } from './brands/KubernetesBrand'
import { LambdaBrand } from './brands/LambdaBrand'
import { NuonBrand } from './brands/NuonBrand'
import { OCIBrand } from './brands/OCIBrand'
import { PulumiBrand } from './brands/PulumiBrand'
import { SlackBrand } from './brands/SlackBrand'
import { TerraformBrand } from './brands/TerraformBrand'
import type { IBrandMark, TBrandTone } from './brands/types'

const BRANDS = {
  AWS: AWSBrand,
  Azure: AzureBrand,
  Docker: DockerBrand,
  GCP: GCPBrand,
  GitHub: GitHubBrand,
  Helm: HelmBrand,
  Kubernetes: KubernetesBrand,
  Lambda: LambdaBrand,
  Nuon: NuonBrand,
  OCI: OCIBrand,
  Pulumi: PulumiBrand,
  Slack: SlackBrand,
  Terraform: TerraformBrand,
} as const

export type TBrandVariant = keyof typeof BRANDS

export interface IBrand extends Omit<IBrandMark, 'tone'> {
  variant: TBrandVariant
  tone?: TBrandTone
}

export const Brand = ({
  variant,
  size = 16,
  tone = 'color',
  ...props
}: IBrand) => {
  const Mark = BRANDS[variant]

  return (
    <Mark
      size={size}
      tone={tone}
      {...props}
      aria-hidden="true"
      focusable="false"
    />
  )
}
