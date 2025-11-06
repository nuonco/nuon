export const runtime = 'nodejs'

import https from 'https'
import http from 'http'
import { NextRequest } from 'next/server'
import { auth0 } from '@/lib/auth'
import { API_URL } from '@/configs/api'
import { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'workflowId' | 'stepId' | 'approvalId'>
) {
  const session = await auth0.getSession()
  const { orgId, workflowId, stepId, approvalId } = await params

  // Parse API_URL for protocol, hostname, and port
  const apiUrlObj = new URL(
    API_URL.startsWith('http') ? API_URL : `http://${API_URL}`
  )
  const isHttps = apiUrlObj.protocol === 'https:'
  const requestModule = isHttps ? https : http
  const hostname = apiUrlObj.hostname
  const port = apiUrlObj.port ? parseInt(apiUrlObj.port) : isHttps ? 443 : 80

  return new Promise<Response>((resolve, reject) => {
    const options: https.RequestOptions = {
      hostname,
      port,
      path: `/v1/workflows/${workflowId}/steps/${stepId}/approvals/${approvalId}/contents`,
      method: 'GET',
      headers: {
        Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': orgId,
        'Accept-Encoding': 'gzip',
      },
    }

    const req = requestModule.request(options, (upstreamRes) => {
      const headers: Record<string, string> = {}
      // Only add string headers (ignore array headers for simplicity)
      Object.entries(upstreamRes.headers).forEach(([key, value]) => {
        if (typeof value === 'string') headers[key] = value
      })

      const chunks: Buffer[] = []

      upstreamRes.on('data', (chunk) => chunks.push(chunk as Buffer))
      upstreamRes.on('end', () => {
        const body = Buffer.concat(chunks)
        resolve(
          new Response(body, {
            status: upstreamRes.statusCode || 200,
            headers,
          })
        )
      })
      upstreamRes.on('error', reject)
    })

    req.on('error', reject)
    req.end()
  })
}
