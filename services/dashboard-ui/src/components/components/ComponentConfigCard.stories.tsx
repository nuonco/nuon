/* eslint-disable react/no-unescaped-entities */
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ComponentConfigCard } from './ComponentConfigCard'
import type { TComponentConfig } from '@/types'

const mockHelmConfig: TComponentConfig = {
  id: 'cfg1234567890abcdef',
  version: 1,
  type: 'helm_chart',
  checksum: 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6',
  build_timeout: '30m',
  deploy_timeout: '1h',
  drift_schedule: '0 */6 * * *',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp123',
  app_config_id: 'cfg123',
  helm: {
    id: 'helm123',
    chart_name: 'nginx-ingress',
    namespace: 'ingress-system',
    storage_driver: 'secret',
    helm_config_json: {
      chart_name: 'nginx-ingress',
      namespace: 'ingress-system',
      storage_driver: 'secret',
      values: {
        'controller.replicaCount': '3',
        'controller.service.type': 'LoadBalancer',
        'controller.metrics.enabled': 'true',
        'defaultBackend.enabled': 'true',
      },
      values_files: [
        `# values.yaml
controller:
  replicaCount: 3
  service:
    type: LoadBalancer
  metrics:
    enabled: true
defaultBackend:
  enabled: true`,
      ],
    },
    connected_github_vcs_config: {
      id: 'vcs123',
      repo: 'myorg/kubernetes-configs',
      directory: 'charts/nginx',
      branch: 'main',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

const mockTerraformConfig: TComponentConfig = {
  id: 'cfg2234567890abcdef',
  version: 2,
  type: 'terraform_module',
  checksum: 'b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7',
  build_timeout: '45m',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp456',
  app_config_id: 'cfg456',
  terraform_module: {
    id: 'tf123',
    version: '1.5.7',
    variables: {
      region: 'us-east-1',
      instance_type: 't3.medium',
      enable_monitoring: 'true',
      vpc_cidr: '10.0.0.0/16',
      environment: 'production',
    },
    variables_files: [
      `# terraform.tfvars
region = "us-east-1"
instance_type = "t3.medium"
enable_monitoring = true
vpc_cidr = "10.0.0.0/16"
environment = "production"`,
    ],
    public_git_vcs_config: {
      id: 'vcs456',
      repo: 'hashicorp/terraform-aws-modules',
      directory: 'modules/vpc',
      branch: 'main',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

const mockKubernetesConfig: TComponentConfig = {
  id: 'cfg3234567890abcdef',
  version: 1,
  type: 'kubernetes_manifest',
  checksum: 'c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8',
  deploy_timeout: '30m',
  drift_schedule: '*/10 * * * *',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp789',
  app_config_id: 'cfg789',
  kubernetes_manifest: {
    id: 'k8s123',
    namespace: 'default',
    manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80`,
    connected_github_vcs_config: {
      id: 'vcs789',
      repo: 'myorg/k8s-manifests',
      directory: 'deployments',
      branch: 'main',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

const mockDockerConfig: TComponentConfig = {
  id: 'cfg4234567890abcdef',
  version: 3,
  type: 'docker_build',
  checksum: 'd4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9',
  build_timeout: '20m',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp101',
  app_config_id: 'cfg101',
  docker_build: {
    id: 'docker123',
    dockerfile: 'Dockerfile',
    target: 'production',
    public_git_vcs_config: {
      id: 'vcs101',
      repo: 'myorg/web-app',
      directory: 'services/api',
      branch: 'main',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

const mockExternalImageConfig: TComponentConfig = {
  id: 'cfg5234567890abcdef',
  version: 1,
  type: 'external_image',
  checksum: 'e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp202',
  app_config_id: 'cfg202',
  external_image: {
    id: 'ext123',
    image_url: 'registry.hub.docker.com/library/redis',
    tag: '7.0-alpine',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

const mockJobConfig: TComponentConfig = {
  id: 'cfg6234567890abcdef',
  version: 1,
  type: 'job',
  checksum: 'f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  component_id: 'cmp303',
  app_config_id: 'cfg303',
  job: {
    id: 'job123',
    image_url: 'myregistry.com/data-processor',
    tag: 'v2.1.0',
    cmd: ['python', 'main.py'],
    args: ['--mode=batch', '--workers=4'],
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
}

export const HelmChart = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Helm Chart Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays Helm chart configuration with chart name, namespace, storage
          driver, and buttons to view values and values files. Includes VCS
          information when available.
        </p>
      </div>

      <ComponentConfigCard config={mockHelmConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Chart name and namespace configuration</li>
          <li>Storage driver display</li>
          <li>"View values" button for key-value pairs</li>
          <li>"View values files" button for YAML content</li>
          <li>Drift schedule with human-readable cron</li>
          <li>VCS repository information</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const TerraformModule = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Terraform Module Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays Terraform module configuration with version and buttons to view
          variables and variables files. Shows both structured variables and raw
          HCL files.
        </p>
      </div>

      <ComponentConfigCard config={mockTerraformConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Terraform version display</li>
          <li>"View variables" button for key-value pairs</li>
          <li>"View variables files" button for HCL content</li>
          <li>Build timeout configuration</li>
          <li>Checksum with click-to-copy</li>
          <li>VCS repository information</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const KubernetesManifest = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">
          Kubernetes Manifest Configuration
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays Kubernetes manifest configuration with namespace and a button
          to view the full YAML manifest. Includes drift detection schedule for
          monitoring configuration changes.
        </p>
      </div>

      <ComponentConfigCard config={mockKubernetesConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Namespace configuration</li>
          <li>"View manifest" button for YAML content</li>
          <li>Drift schedule showing frequency of checks</li>
          <li>Deploy timeout display</li>
          <li>VCS repository information</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const DockerBuild = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Docker Build Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays Docker build configuration showing the Dockerfile name and
          build target. The actual Dockerfile content is stored in the VCS
          repository.
        </p>
      </div>

      <ComponentConfigCard config={mockDockerConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Dockerfile name (not full path)</li>
          <li>Build target for multi-stage builds</li>
          <li>Build timeout configuration</li>
          <li>VCS repository showing Dockerfile location</li>
          <li>No modal buttons (content is in repository)</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const ExternalImage = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">External Image Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays external container image configuration with registry URL and
          tag. Used for images that are built externally and pulled from a
          registry.
        </p>
      </div>

      <ComponentConfigCard config={mockExternalImageConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Image URL showing registry and image name</li>
          <li>Image tag/version</li>
          <li>Minimal configuration (no builds or VCS)</li>
          <li>No modal buttons (external image source)</li>
          <li>Checksum for configuration tracking</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const JobConfiguration = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">Job Configuration</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Displays job configuration for batch processing or scheduled tasks.
          Shows container image, command, and arguments used to run the job.
        </p>
      </div>

      <ComponentConfigCard config={mockJobConfig} />

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Key Features:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>Image URL and tag for job container</li>
          <li>Command array showing execution command</li>
          <li>Arguments for job parameterization</li>
          <li>No VCS (uses existing container image)</li>
          <li>Suitable for Lambda-like workloads</li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)

export const AllConfigTypes = () => (
  <SurfacesProvider>
    <div className="space-y-6">
      <div className="space-y-3">
        <h3 className="text-lg font-semibold">All Component Configuration Types</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Overview of all supported component configuration types showing how each
          type displays different fields and actions based on the component's
          nature.
        </p>
      </div>

      <div className="space-y-6">
        <ComponentConfigCard config={mockHelmConfig} />
        <ComponentConfigCard config={mockTerraformConfig} />
        <ComponentConfigCard config={mockKubernetesConfig} />
        <ComponentConfigCard config={mockDockerConfig} />
        <ComponentConfigCard config={mockExternalImageConfig} />
        <ComponentConfigCard config={mockJobConfig} />
      </div>

      <div className="text-sm text-gray-600 dark:text-gray-400 mt-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
        <strong>Component Types:</strong>
        <ul className="mt-2 space-y-1 list-disc list-inside">
          <li>
            <strong>Helm Chart:</strong> Kubernetes package manager with values
            customization
          </li>
          <li>
            <strong>Terraform Module:</strong> Infrastructure as code with
            variable configuration
          </li>
          <li>
            <strong>Kubernetes Manifest:</strong> Raw K8s YAML configuration files
          </li>
          <li>
            <strong>Docker Build:</strong> Container images built from Dockerfiles
          </li>
          <li>
            <strong>External Image:</strong> Pre-built images from registries
          </li>
          <li>
            <strong>Job:</strong> Batch processing or scheduled task containers
          </li>
        </ul>
      </div>
    </div>
  </SurfacesProvider>
)
