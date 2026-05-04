import { DateTime } from 'luxon'
import { RetryNowButton as RetryNowButtonComponent } from './RetryNow'

export default {
  title: 'Workflows/RetryNow',
}

export const CountingDown = () => (
  <RetryNowButtonComponent
    retryNotBeforeAt={DateTime.now().plus({ minutes: 2, seconds: 30 }).toISO()!}
    isPending={false}
    onTrigger={() => {}}
  />
)

export const AlmostReady = () => (
  <RetryNowButtonComponent
    retryNotBeforeAt={DateTime.now().plus({ seconds: 5 }).toISO()!}
    isPending={false}
    onTrigger={() => {}}
  />
)

export const ReadyNow = () => (
  <RetryNowButtonComponent
    retryNotBeforeAt={DateTime.now().toISO()!}
    isPending={false}
    onTrigger={() => {}}
  />
)

export const Pending = () => (
  <RetryNowButtonComponent
    retryNotBeforeAt={DateTime.now().plus({ minutes: 1 }).toISO()!}
    isPending={true}
    onTrigger={() => {}}
  />
)
