import { delay, http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
import { faker } from '@faker-js/faker'
import {
  handlers,
  getGetBuild200Response,
  getGetOrg200Response,
  getGetInstall200Response,
  getGetInstallComponents200Response,
  getGetInstallDeploy200Response,
  getGetInstallSandboxRuns200Response,
} from './mock-api-handlers'

const baseURL = 'https://api.nuon.co'

export const customHandlers = [
  http.get(`${baseURL}/v1/workflows/pending-approvals`, () => {
    return HttpResponse.json(
      Array.from({ length: 3 }, () => ({
        id: faker.lorem.slug(1),
        created_at: faker.date.past(),
        updated_at: faker.date.past(),
        workflow_step_id: faker.lorem.slug(1),
        type: faker.helpers.arrayElement(['noop', 'approve-all', 'terraform_plan']),
      })),
      { status: 200 }
    )
  }),
]

export const nextProxyHandlers = [
  http.get('/api/:orgId', async () => {
    await delay(300)
    return HttpResponse.json(getGetOrg200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstall200Response(), {
      status: 200,
    })
  }),

  http.get(
    '/api/:orgId/installs/:installId/components/:installComponentId',
    async () => {
      await delay(300)
      return HttpResponse.json(getGetInstallComponents200Response()[0], {
        status: 200,
      })
    }
  ),

  http.get('/api/:orgId/components/:componentId/builds/:buildId', async () => {
    await delay(300)
    return HttpResponse.json(getGetBuild200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId/deploys/:deployId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstallDeploy200Response(), {
      status: 200,
    })
  }),

  http.get('/api/:orgId/installs/:installId/runs/:runId', async () => {
    await delay(300)
    return HttpResponse.json(getGetInstallSandboxRuns200Response()[0], {
      status: 200,
    })
  }),
]

export const server = setupServer(...customHandlers, ...handlers)
