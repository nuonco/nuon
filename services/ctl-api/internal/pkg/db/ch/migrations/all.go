package migrations

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"

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
	}
}
