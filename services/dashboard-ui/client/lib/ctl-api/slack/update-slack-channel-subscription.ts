import type { Interests } from '@/components/interests/types'
import type { SubscriptionMatch } from '@/components/match/types'
import { api } from '@/lib/api'
import type { TSlackChannelSubscription } from '@/types'

// Body for PATCH /v1/orgs/{org_id}/slack/channel-subscriptions/{sub_id}.
//
// All fields are optional — pass only the fields you want to change.
// `match` follows PUT semantics on the server: include it (with `null`
// for org-wide) when you want to change the routing predicate, omit it
// to leave the existing match alone. Updating the match may collide with
// the unique index on (team_id, channel_id, org_link_id, match_canonical);
// the server returns 409 in that case.
export interface UpdateSlackChannelSubscriptionBody {
  channel_id?: string
  channel_name?: string
  match?: SubscriptionMatch | null
  interests?: Interests
}

export const updateSlackChannelSubscription = ({
  body,
  orgId,
  subId,
}: {
  body: UpdateSlackChannelSubscriptionBody
  orgId: string
  subId: string
}) =>
  api<TSlackChannelSubscription>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/${orgId}/slack/channel-subscriptions/${subId}`,
  })
