export const runtime = 'nodejs'

import https from 'https'
import { NextRequest } from 'next/server'
import { API_URL, auth0 } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

function stripProtocol(url: string) {
  // Remove protocol, trailing slash, and path/query if present
  return url.replace(/^https?:\/\//, '').split('/')[0]
}

export async function GET(
  request: NextRequest,
  { params }: TRouteRes<'org-id' | 'workflow-id' | 'step-id' | 'approval-id'>
) {
  const session = await auth0.getSession()
  const {
    ['org-id']: orgId,
    ['workflow-id']: workflowId,
    ['step-id']: stepId,
    ['approval-id']: approvalId,
  } = await params

  return new Promise<Response>((resolve, reject) => {
    const options: https.RequestOptions = {
      hostname: stripProtocol(API_URL),
      path: `/v1/workflows/${workflowId}/steps/${stepId}/approvals/${approvalId}/contents`,
      method: 'GET',
      headers: {
        Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': orgId,
        'Accept-Encoding': 'gzip',
      },
    }

    const req = https.request(options, (upstreamRes) => {
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
