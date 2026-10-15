package componentsyncimage

const (
	checkpointPreparation = "sync-image-preparation"
	checkpointSyncImage   = "sync-image"
)

type preparationCheckpoint struct {
	DeployID    string `json:"deploy_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}
