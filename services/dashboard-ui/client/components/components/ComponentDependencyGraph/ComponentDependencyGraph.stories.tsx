import { ComponentDependencyGraph } from './ComponentDependencyGraph'

export default {
  title: 'Components/ComponentDependencyGraph',
}

const basePath = '/org-1/apps/app-1/components'

export const WithDepsAndDependents = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      current={{ id: 'comp-1', name: 'api-server', type: 'helm_chart' }}
      dependencies={[
        { id: 'comp-2', name: 'vpc', type: 'terraform_module' },
        { id: 'comp-3', name: 'database', type: 'terraform_module' },
      ]}
      dependents={[
        { id: 'comp-4', name: 'frontend', type: 'docker_build' },
        { id: 'comp-5', name: 'worker', type: 'helm_chart' },
        { id: 'comp-6', name: 'cron-jobs', type: 'kubernetes_manifest' },
      ]}
      basePath={basePath}
    />
  </div>
)

export const DependenciesOnly = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      current={{ id: 'comp-1', name: 'frontend', type: 'docker_build' }}
      dependencies={[
        { id: 'comp-2', name: 'api-server', type: 'helm_chart' },
      ]}
      dependents={[]}
      basePath={basePath}
    />
  </div>
)

export const DependentsOnly = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      current={{ id: 'comp-1', name: 'vpc', type: 'terraform_module' }}
      dependencies={[]}
      dependents={[
        { id: 'comp-2', name: 'eks-cluster', type: 'terraform_module' },
        { id: 'comp-3', name: 'database', type: 'terraform_module' },
      ]}
      basePath={basePath}
    />
  </div>
)

export const SingleNode = () => (
  <div style={{ width: 600, height: 400 }}>
    <ComponentDependencyGraph
      current={{ id: 'comp-1', name: 'standalone', type: 'helm_chart' }}
      dependencies={[]}
      dependents={[]}
      basePath={basePath}
    />
  </div>
)
