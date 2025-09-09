import { NextRequest, NextResponse } from 'next/server'
import { ADMIN_API_URL } from '@/configs/api'

const BACKEND_BASE_URL = ADMIN_API_URL || 'http://localhost:8082'

export async function POST(req: NextRequest) {
  const body = await req.text()

  const headers = new Headers(req.headers)
  headers.delete('host')
  headers.set('content-type', 'application/json')

  const response = await fetch(
    `${BACKEND_BASE_URL}/v1/general/temporal-codec/decode`,
    {
      method: 'POST',
      body,
      headers,
    }
  )

  // Forward response body as text, and clean up headers
  const resBody = await response.text()
  const proxyHeaders = new Headers(response.headers)
  proxyHeaders.set('content-type', 'application/json')
  proxyHeaders.delete('content-length')
  proxyHeaders.delete('transfer-encoding')
  proxyHeaders.delete('content-encoding')

  return new NextResponse(resBody, {
    status: response.status,
    headers: proxyHeaders,
  })
}
