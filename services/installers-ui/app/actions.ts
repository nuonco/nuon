'use server'

import { installRequestBody } from './util'

const API_URL = process.env.NUON_API_URL

export async function getOrg(orgId): Promise<Record<string, any>> {
  console.debug(`[getOrg] orgId=${orgId}`)
  const res = await fetch(`${API_URL}/v1/orgs/${orgId}`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
    },
  })

  return res.json()
}

export async function getInstaller(orgId): Promise<Record<string, any>> {
  console.debug(`[getInstaller] orgId=${orgId}`)
  // TODO: determine where the slug/installer-id comes from
  const res = await fetch(`${API_URL}/v1/installers`, {
    method: 'get',
    cache: 'no-store',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
  })
  let installers = await res.json()
  if (installers.length < 1) {
    // TODO: raise and bubble an error
    return {}
  }
  let firstInstaller = installers[0]
  return firstInstaller
}

export async function getAppBySlug(
  slug: string,
  orgId: string
): Promise<Record<string, any>> {
  console.debug(`[getAppBySlug] ${slug} ${orgId}`)
  const res = await fetch(`${API_URL}/v1/apps/${slug}`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
  })

  return res.json()
}

export async function createInstall(
  app: Record<string, any>,
  formData: FormData,
  orgId: string
) {
  const input = installRequestBody(app, formData)

  const res = await fetch(`${API_URL}/v1/apps/${app?.id}/installs`, {
    body: JSON.stringify(input),
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
    method: 'POST',
  })

  return await res.json()
}

export async function getInstall(
  id: string,
  orgId: string
): Promise<Record<string, any>> {
  const res = await fetch(`${API_URL}/v1/installs/${id}`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
  })

  // work-around for API bug
  let json = await res.json()
  if (Array.isArray(json)) {
    json = json[0]
  }

  return json
}

export async function updateInstall(
  id: string,
  app: Record<string, any>,
  formData: FormData,
  orgId: string
) {
  const input = installRequestBody(app, formData)

  const res = await fetch(`${API_URL}/v1/installs/${id}`, {
    body: JSON.stringify(input),
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
    method: 'PATCH',
  })

  return await res.json()
}

export async function reprovisionInstall(
  id: string,
  orgId: string
): Promise<Record<string, any>> {
  const res = await fetch(`${API_URL}/v1/installs/${id}/reprovision`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
    method: 'POST',
  })

  return res.json()
}

export async function deployComponents(
  id: string,
  orgId: string
): Promise<Record<string, any>> {
  const res = await fetch(
    `${API_URL}/v1/installs/${id}/components/deploy-all`,
    {
      cache: 'no-store',
      headers: {
        Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
        'X-Nuon-Org-ID': orgId,
      },
      method: 'POST',
    }
  )

  return res.json()
}

export async function updateInputs(
  id: string,
  app: Record<string, any>,
  formData: FormData,
  orgId: string
): Promise<Record<string, any>> {
  const input = installRequestBody(app, formData)

  const res = await fetch(`${API_URL}/v1/installs/${id}/inputs`, {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${process?.env?.NUON_API_TOKEN}`,
      'X-Nuon-Org-ID': orgId,
    },
    method: 'POST',
    body: JSON.stringify({ inputs: input.inputs }),
  })

  return res.json()
}

export async function redeployInstall(
  id: string,
  app: Record<string, any>,
  formData: FormData,
  orgId: string
) {
  const reqBody = installRequestBody(app, formData)

  const updateRes = await updateInstall(id, app, formData, orgId)
  if (updateRes.error) {
    return updateRes
  }

  if (Object.keys(reqBody.inputs).length > 0) {
    const inputsRes = await updateInputs(id, app, formData, orgId)
    if (inputsRes.error) {
      return inputsRes
    }
  }

  const reproRes = await reprovisionInstall(id, orgId)
  if (reproRes.error) {
    return reproRes
  }

  return await deployComponents(id, orgId)
}
