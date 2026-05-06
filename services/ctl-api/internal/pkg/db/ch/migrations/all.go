package migrations

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"

func (m *Migrations) All() []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "01-create-latest-runner-heart-beats",
			Fn:   m.Migration001LatestRunnerHeartBeats,
		},
		{
			Name: "02-create-latest-runner-heart-beats-mv-v1",
			Fn:   m.Migration002LatestRunnerHeartBeatsMaterializedViewV1,
		},
		{
			Name: "03-add-process-id-to-heartbeats",
			Fn:   m.Migration003AddProcessIDToHeartBeats,
		},
		{
			Name: "04-recreate-heartbeat-sort-keys",
			Fn:   m.Migration004RecreateHeartbeatSortKeys,
		},
		{
			Name: "05-recreate-runner-health-check-sort-key",
			Fn:   m.Migration005RecreateRunnerHealthCheckSortKey,
		},
		{
			Name: "06-fix-otel-log-attr-indexes",
			Fn:   m.Migration006FixOtelLogAttrIndexes,
		},
	}
}
