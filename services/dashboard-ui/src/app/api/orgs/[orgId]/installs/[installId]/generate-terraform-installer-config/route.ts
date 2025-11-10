import { type NextRequest, NextResponse } from 'next/server'
import { API_URL } from '@/configs/api'
import { auth0 } from '@/lib/auth'
import type { TRouteProps } from '@/types'

// API route for downloading Terraform installer config

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params

  try {
    const session = await auth0.getSession()

    const response = await fetch(
      `${API_URL}/v1/installs/${installId}/generate-terraform-installer-config`,
      {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
          'X-Nuon-Org-ID': orgId,
        },
      }
    )

    if (!response.ok) {
      const errorText = await response.text()
      console.error('Failed to fetch terraform config:', {
        status: response.status,
        statusText: response.statusText,
        url: `${API_URL}/v1/installs/${installId}/generate-terraform-installer-config`,
        orgId,
        installId,
        errorBody: errorText
      })
      return new Response(`Failed to generate terraform installer config: ${response.status} ${response.statusText}`, {
        status: response.status,
      })
    }

    // Get the response as binary data
    const configData = await response.arrayBuffer()

    // Forward the response with proper headers for file download
    return new Response(configData, {
      status: 200,
      headers: {
        'Content-Type': 'application/octet-stream',
        'Content-Disposition': 'attachment; filename="terraform.tfvars"',
      },
    })
  } catch (error) {
    console.error('Error generating terraform installer config:', error)
    return new Response('Internal server error', { status: 500 })
  }
}