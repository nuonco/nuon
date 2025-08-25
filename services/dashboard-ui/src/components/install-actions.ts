'use server'

import { revalidatePath } from 'next/cache'
import {
  deployComponents as deployAllComponents,
  reprovisionInstall as reprovision,
  reprovisionSandbox as reprovisionSBox,
  deployComponentBuild as deployComponentByBuildId,
  teardownInstallComponents,
  updateInstall as patchInstall,
  forgetInstall as forget,
  installManagedByUI,
} from '@/lib'
import { API_URL, nueMutateData, getFetchOpts } from '@/utils'
import type { TInstall } from '@/types'

interface IReprovisionInstall {
  installId: string
  orgId: string
  continueOnError?: boolean
  planOnly?: boolean
}

export async function reprovisionInstall({
  continueOnError = false,
  installId,
  orgId,
  planOnly,
}: IReprovisionInstall) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/reprovision`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      error_behavior: continueOnError ? 'continue' : 'abort',
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off install reprovision')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

export async function reprovisionSandbox({
  continueOnError = false,
  installId,
  orgId,
  planOnly,
}: IReprovisionInstall) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/reprovision-sandbox`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      error_behavior: continueOnError ? 'continue' : 'abort',
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off sandbox reprovision')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

export async function syncSecrets({
  continueOnError = false,
  installId,
  orgId,
  planOnly,
}: IReprovisionInstall) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/sync-secrets`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      error_behavior: continueOnError ? 'continue' : 'abort',
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off secrets sync')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IDeployComponents {
  continueOnError?: boolean
  installId: string
  orgId: string
  planOnly?: boolean
}

export async function deployComponents({
  continueOnError = false,
  installId,
  orgId,
  planOnly,
}: IDeployComponents) {
  const res = fetch(
    `${API_URL}/v1/installs/${installId}/components/deploy-all`,
    {
      ...(await getFetchOpts(orgId)),

      body: JSON.stringify({
        error_behavior: continueOnError ? 'continue' : 'abort',
        plan_only: planOnly,
      }),
      method: 'POST',
    }
  )
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off components deployment')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IDeployComponentBuild {
  buildId: string
  installId: string
  orgId: string
  continueOnError?: boolean
  deployDeps?: boolean
  planOnly?: boolean
}

export async function deployComponentBuild({
  buildId,
  continueOnError = false,
  deployDeps = false,
  installId,
  orgId,
  planOnly,
}: IDeployComponentBuild) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/deploys`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      // TODO(nnnnat): assuming we will want to enable this soon
      //error_behavior: continueOnError ? 'continue' : 'abort',
      build_id: buildId,
      deploy_dependents: deployDeps,
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off build deploy')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IRevalidateInstallData {
  installId: string
  orgId: string
}

export async function revalidateInstallData({
  orgId,
  installId,
}: IRevalidateInstallData) {
  revalidatePath(`/${orgId}/installs/${installId}`)
}

interface ITeardownAllComponents {
  continueOnError?: boolean
  installId: string
  orgId: string
  planOnly?: boolean
}

export async function teardownAllComponents({
  continueOnError = false,
  installId,
  orgId,
  planOnly,
}: ITeardownAllComponents) {
  const res = fetch(
    `${API_URL}/v1/installs/${installId}/components/teardown-all`,
    {
      ...(await getFetchOpts(orgId)),
      body: JSON.stringify({
        error_behavior: continueOnError ? 'continue' : 'abort',
        plan_only: planOnly,
      }),
      method: 'POST',
    }
  )
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off component delete')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IUpdateInstall {
  installId: string
  orgId: string
  formData: FormData
}

export async function updateInstall({
  installId,
  orgId,
  formData: fd,
}: IUpdateInstall) {
  const formData = Object.fromEntries(fd)

  const inputs = Object.keys(formData).reduce((acc, key) => {
    if (key.includes('inputs:')) {
      let value: any = formData[key]
      if (value === 'on' || value === 'off') {
        value = Boolean(value === 'on').toString()
      }

      acc[key.replace('inputs:', '')] = value
    }

    return acc
  }, {})

  if (Object.keys(inputs)?.length > 0) {
    const res = fetch(`${API_URL}/v1/installs/${installId}/inputs`, {
      ...(await getFetchOpts(orgId)),
      body: JSON.stringify({ inputs }),
      method: 'PATCH',
    })
      .then((r) => {
        if (!r.ok) {
          throw new Error('Unable to update inputs')
        } else {
          return r
        }
      })
      .catch((err) => {
        throw new Error(err)
      })

    return (await res).headers.get('x-nuon-install-workflow-id')
  }
}

interface IForgetInstall {
  installId: string
  orgId: string
}

export async function forgetInstall(params: IForgetInstall) {
  return forget(params)
}

interface IDeleteComponents {
  continueOnError?: boolean
  installId: string
  orgId: string
  force?: boolean
  planOnly?: boolean
}

export async function deleteComponents({
  continueOnError = false,
  installId,
  orgId,
  force = false,
  planOnly,
}: IDeleteComponents) {
  const res = fetch(
    `${API_URL}/v1/installs/${installId}/components/teardown-all`,
    {
      ...(await getFetchOpts(orgId)),
      body: JSON.stringify({
        error_behavior: force ? 'continue' : 'abort',
        plan_only: planOnly,
      }),
      method: 'POST',
    }
  )
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off components delete')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IDeleteComponent {
  continueOnError?: boolean
  componentId: string
  installId: string
  orgId: string
  force?: boolean
  planOnly?: boolean
}

export async function deleteComponent({
  continueOnError = false,
  componentId,
  installId,
  orgId,
  force = false,
  planOnly,
}: IDeleteComponent) {
  let error = null
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components/${componentId}/teardown`,
    {
      ...(await getFetchOpts(orgId)),
      body: JSON.stringify({
        error_behavior: force ? 'continue' : 'abort',
        plan_only: planOnly,
      }),
      method: 'POST',
    }
  )
    .then(async (r) => {
      if (!r.ok) {
        error = await r?.json()
      } else {
        return r
      }
    })
    .catch((err) => {
      error = err
    })

  return {
    data: res ? res.headers.get('x-nuon-install-workflow-id') : null,
    error,
  }
}

interface IDeleteInstall {
  installId: string
  orgId: string
  force?: boolean
  planOnly?: boolean
}

export async function deleteInstall({
  installId,
  orgId,
  force = false,
  planOnly,
}: IDeleteInstall) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/deprovision`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      error_behavior: force ? 'continue' : 'abort',
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off install deprovision')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IDeprovisionSandbox extends IDeleteComponents { }

export async function deprovisionSandbox({
  continueOnError = false,
  installId,
  orgId,
  force = false,
  planOnly,
}: IDeprovisionSandbox) {
  const res = fetch(`${API_URL}/v1/installs/${installId}/deprovision-sandbox`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({
      error_behavior: force ? 'continue' : 'abort',
      plan_only: planOnly,
    }),
    method: 'POST',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to kick off components delete')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).headers.get('x-nuon-install-workflow-id')
}

interface IApproveWorklowStep {
  orgId: string
  workflowId: string
  stepId: string
  approvalId: string
  responseType: 'approve' | 'deny' | 'retry'
}

export async function approveWorkflowStep({
  approvalId,
  orgId,
  responseType,
  stepId,
  workflowId,
}: IApproveWorklowStep) {
  return nueMutateData({
    orgId,
    path: `install-workflows/${workflowId}/steps/${stepId}/approvals/${approvalId}/response`,
    body: {
      note: '',
      response_type: responseType,
    },
  })
}

interface ICreateInstallConfig {
  approvalOption: 'approve-all' | 'prompt'
  installId: string
  orgId: string
}

export async function createInstallConfig({
  approvalOption,
  installId,
  orgId,
}: ICreateInstallConfig) {
  return nueMutateData({
    orgId,
    path: `installs/${installId}/configs`,
    body: {
      approval_option: approvalOption,
    },
  })
}

interface IUpdateInstallConfig {
  approvalOption: 'approve-all' | 'prompt'
  configId: string
  installId: string
  orgId: string
}

export async function updateInstallConfig({
  approvalOption,
  configId,
  installId,
  orgId,
}: IUpdateInstallConfig) {
  return nueMutateData({
    orgId,
    path: `installs/${installId}/configs/${configId}`,
    method: 'PATCH',
    body: {
      approval_option: approvalOption,
    },
  })
}

interface IUpdateInstallManagedBy {
  installId: string
  managedBy?: string
  orgId: string
}

export async function updateInstallManagedBy({
  installId,
  managedBy,
  orgId,
}: IUpdateInstallManagedBy) {
  if (managedBy && managedBy === installManagedByUI) {
    return true
  }
  const res = fetch(`${API_URL}/v1/installs/${installId}`, {
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({ metadata: { managed_by: installManagedByUI } }),
    method: 'PATCH',
  })
    .then((r) => {
      if (!r.ok) {
        throw new Error('Unable to mark install managed by UI')
      } else {
        return r
      }
    })
    .catch((err) => {
      throw new Error(err)
    })

  return (await res).ok
}

interface IRetryWorklow {
  orgId: string
  workflowId: string
  stepId: string
  op: 'retry-step' | 'skip-step'
}

export async function retryWorkflow({
  orgId,
  workflowId,
  stepId,
  op,
}: IRetryWorklow) {
  let path = `workflows/${workflowId}/retry`
  return nueMutateData({
    orgId,
    path,
    body: {
      step_id: stepId,
      operation: op
    },
  })
}
