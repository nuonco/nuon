'use server'

import type { TAccount } from '@/types'
import { nueQueryData } from '@/utils'
import { getFetchOpts } from '@/utils/get-fetch-opts'
import { API_URL } from '@/configs/api'

export async function getCurrentAccount(): Promise<TAccount | null> {
  try {
    const { data: account, error } = await nueQueryData<TAccount>({
      path: 'account',
    })

    if (error) {
      console.error('Failed to fetch current account:', error)
      return null
    }

    return account || null
  } catch (err) {
    console.error('Error fetching current account:', err)
    return null
  }
}

export async function completeUserJourney(journeyName: string): Promise<boolean> {
  try {
    const response = await fetch(
      `${API_URL}/v1/account/user-journeys/${journeyName}/complete`,
      {
        ...(await getFetchOpts()),
        method: 'POST',
      }
    )

    if (!response.ok) {
      const errorText = await response.text()
      console.error(
        `Failed to complete ${journeyName} journey:`,
        response.status,
        errorText
      )
      return false
    }

    return true
  } catch (err) {
    console.error(`Error completing ${journeyName} journey:`, err)
    return false
  }
}

export async function completeEvaluationJourney(): Promise<boolean> {
  return completeUserJourney('evaluation')
}

