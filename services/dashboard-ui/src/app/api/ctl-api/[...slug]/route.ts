import { NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'

// Handles both Swagger UI/static asset proxying (.js, .css, .ico, .html, .png, .svg, etc) and API requests.
// For asset requests, sets the correct content-type. For API requests, passes through everything.
async function buildTargetUrl(p: Promise<{ slug?: string[] }>, req: NextRequest) {
  const params = (await p) || {}
  // If no slug, serve "/" (root of API_URL), else join slug for asset/api path
  const urlPath = Array.isArray(params.slug) && params.slug.length > 0 ? '/' + params.slug.join('/') : ''
  return `${API_URL}${urlPath}${req.nextUrl.search}`
}

async function proxyRequest(
  method: string,
  req: NextRequest,
  paramsPromise: Promise<{ slug?: string[] }>
) {
  const targetUrl = await buildTargetUrl(paramsPromise, req)

  // Clone all headers except compression
  const headers = { ...Object.fromEntries(req.headers) }
  delete headers['accept-encoding']

  let fetchInit: RequestInit = {
    method,
    headers,
  }

  // Pass through body for mutating requests
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

  // Clone response headers, strip problematic encodings
  const responseHeaders = new Headers(res.headers)
  responseHeaders.delete('content-encoding')
  responseHeaders.delete('transfer-encoding')
  responseHeaders.delete('content-length')

  // Set content-type for known Swagger/static assets
  if (targetUrl.endsWith('.js')) {
    responseHeaders.set('content-type', 'application/javascript')
  } else if (targetUrl.endsWith('.css')) {
    responseHeaders.set('content-type', 'text/css')
  } else if (targetUrl.endsWith('.ico')) {
    responseHeaders.set('content-type', 'image/x-icon')
  } else if (targetUrl.endsWith('.html')) {
    responseHeaders.set('content-type', 'text/html; charset=utf-8')
  } else if (targetUrl.endsWith('.png')) {
    responseHeaders.set('content-type', 'image/png')
  } else if (targetUrl.endsWith('.svg')) {
    responseHeaders.set('content-type', 'image/svg+xml')
  } else if (targetUrl.endsWith('.json')) {
    responseHeaders.set('content-type', 'application/json')
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
