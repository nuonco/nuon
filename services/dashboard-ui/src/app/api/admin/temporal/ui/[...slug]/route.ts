import { NextRequest, NextResponse } from 'next/server'
import { ADMIN_TEMPORAL_UI_URL } from '@/configs/api'

async function buildTargetUrl(p, req: NextRequest) {
  const params = await p
  const urlPath =
    params.slug && params.slug.length > 0 ? '/' + params.slug.join('/') : ''
  return `${ADMIN_TEMPORAL_UI_URL}${urlPath}${req.nextUrl.search}`
}

async function proxyRequest(
  method: string,
  req: NextRequest,
  paramsPromise: Promise<{ slug?: string[] }>
) {
  const params = await paramsPromise
  const targetUrl = await buildTargetUrl(params, req)

  const headers = { ...Object.fromEntries(req.headers) }
  delete headers['accept-encoding']

  let fetchInit: RequestInit = {
    method,
    headers,
  }

  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
    const contentType = req.headers.get('content-type') || ''
    if (contentType.includes('application/json')) {
      fetchInit.body = JSON.stringify(await req.json())
    } else {
      fetchInit.body = await req.text()
    }
  }

  const res = await fetch(targetUrl, fetchInit)
  const body = await res.arrayBuffer()

  const responseHeaders = new Headers(res.headers)
  responseHeaders.delete('content-encoding')
  responseHeaders.delete('transfer-encoding')
  responseHeaders.delete('content-length')

  // Set content-type for known assets
  if (targetUrl.endsWith('.js')) {
    responseHeaders.set('content-type', 'application/javascript')
  }
  if (targetUrl.endsWith('.css')) {
    responseHeaders.set('content-type', 'text/css')
  }
  if (targetUrl.endsWith('.ico')) {
    responseHeaders.set('content-type', 'image/x-icon')
  }

  return new NextResponse(body, {
    status: res.status,
    headers: responseHeaders,
  })
}

export async function GET(req: NextRequest, context: any) {
  return proxyRequest('GET', req, context.params)
}
export async function POST(req: NextRequest, context: any) {
  return proxyRequest('POST', req, context.params)
}
export async function PUT(req: NextRequest, context: any) {
  return proxyRequest('PUT', req, context.params)
}
export async function DELETE(req: NextRequest, context: any) {
  return proxyRequest('DELETE', req, context.params)
}
export async function PATCH(req: NextRequest, context: any) {
  return proxyRequest('PATCH', req, context.params)
}
export async function OPTIONS(req: NextRequest, context: any) {
  return proxyRequest('OPTIONS', req, context.params)
}
export async function HEAD(req: NextRequest, context: any) {
  return proxyRequest('HEAD', req, context.params)
}
