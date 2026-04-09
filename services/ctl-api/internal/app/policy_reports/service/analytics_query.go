package service

import "time"

type analyticsFilter struct {
	OrgID     string
	AppID     string
	Start     time.Time
	End       time.Time
	InstallID string
	PolicyID  string
}

func buildBaseWhereClause(f analyticsFilter) (string, []any) {
	query := `WHERE org_id = ? AND app_id = ? AND evaluated_at BETWEEN ? AND ?`
	params := []any{f.OrgID, f.AppID, f.Start, f.End}

	if f.InstallID != "" {
		query += " AND install_id = ?"
		params = append(params, f.InstallID)
	}
	if f.PolicyID != "" {
		query += " AND policy_id = ?"
		params = append(params, f.PolicyID)
	}

	return query, params
}

func buildTimeseriesSelectClauses(interval timeInterval, groupBy string) (selectCols, groupCols, orderCols string) {
	bucketExpr := interval.BucketExpr("evaluated_at")
	countCols := `count()                   AS evaluations,
		countIf(outcome = 'deny') AS denies,
		countIf(outcome = 'warn') AS warns,
		countIf(outcome = 'pass') AS passes`

	if groupBy != "" {
		selectCols = bucketExpr + " AS bucket,\n\t\t" + groupBy + " AS group_key,\n\t\t" + countCols
		groupCols = "bucket, group_key"
		orderCols = "bucket, group_key"
	} else {
		selectCols = bucketExpr + " AS bucket,\n\t\t" + countCols
		groupCols = "bucket"
		orderCols = "bucket"
	}

	return selectCols, groupCols, orderCols
}
