import { type NextRequest, NextResponse } from 'next/server'
import { getAvailableRoles } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>,
) {
  const { installId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const operationType = searchParams.get('operation_type')
  const principalType = searchParams.get('principal_type')

  if (!operationType || !principalType) {
    return NextResponse.json(
      { error: 'operation_type and principal_type are required' },
      { status: 400 }
    )
  }

  const response = await getAvailableRoles({
    installId,
    operationType: operationType as any,
    principalType: principalType as any,
    orgId,
  })

  return NextResponse.json(response)
}
