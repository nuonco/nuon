import { type NextRequest, NextResponse } from 'next/server'
import { getAPIHealth } from '@/lib'

export async function GET(_: NextRequest) {
  const response = await getAPIHealth()
  return NextResponse.json(response)
}
