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

const VIEW_BOXES: Record<TBrandVariant, string> = {
  AWS: '1.67 1.1 300.67 179.8',
  Azure: '5.92 6.54 82.55 83.54',
  Docker: '0 3.39 24 17.22',
  GCP: '0 0 256.04 206.04',
  GitHub: '0 8 496 483.61',
  Helm: '1.62 0 20.75 24',
  Kubernetes: '0 0.36 24 23.29',
  Lambda: '0.55 0 22.91 24',
  Nuon: '0 0 20.23 28',
  OCI: '0 0 24 24',
  Pulumi: '0.65 0 22.69 24',
  Slack: '0 0 24 24',
  Terraform: '1.44 0 21.12 24',
}

export const Brand = ({
  variant,
  size = 16,
  tone = 'color',
  style,
  ...props
}: IBrand) => {
  const Mark = BRANDS[variant]
  const viewBox = VIEW_BOXES[variant]
  const [, , boxWidth, boxHeight] = viewBox.split(' ').map(Number)
  const aspect = boxWidth / boxHeight
  const heightScale = aspect > 1 ? 1 / Math.sqrt(aspect) : 1
  const length = (scale: number) =>
    typeof size === 'number' ? `${size * scale}px` : `calc(${size} * ${scale})`

  return (
    <Mark
      tone={tone}
      viewBox={viewBox}
      width={undefined}
      height={undefined}
      style={{
        width: length(heightScale * aspect),
        height: length(heightScale),
        ...style,
      }}
      {...props}
      aria-hidden="true"
      focusable="false"
    />
  )
}
