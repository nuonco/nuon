import { NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'

async function buildTargetUrl(
  paramsPromise: Promise<{ slug?: string[] }>,
  req: NextRequest
) {
  // Always await params!
  const params = (await paramsPromise) || {}
  // If no slug, serve /docs/index.html (main swagger page)
  let urlPath = '/docs/index.html'
  if (Array.isArray(params.slug) && params.slug.length > 0) {
    urlPath = '/docs/' + params.slug.join('/')
  }
  return `${API_URL}${urlPath}${req.nextUrl.search}`
}

async function proxyRequest(
  method: string,
  req: NextRequest,
  paramsPromise: Promise<{ slug?: string[] }>
) {
  const targetUrl = await buildTargetUrl(paramsPromise, req)

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

  // Set correct content-type for HTML and other assets
  const responseHeaders = new Headers(res.headers)
  responseHeaders.delete('content-encoding')
  responseHeaders.delete('transfer-encoding')
  responseHeaders.delete('content-length')

  if (targetUrl.endsWith('.js')) {
    responseHeaders.set('content-type', 'application/javascript')
  } else if (targetUrl.endsWith('.css')) {
    responseHeaders.set('content-type', 'text/css')
  } else if (targetUrl.endsWith('.ico')) {
    responseHeaders.set('content-type', 'image/x-icon')
  } else if (targetUrl.endsWith('.html')) {
    responseHeaders.set('content-type', 'text/html; charset=utf-8')
  }

  return new NextResponse(body, {
    status: res.status,
    headers: responseHeaders,
  })
}

// Always pass context.params as a promise to proxyRequest!
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
