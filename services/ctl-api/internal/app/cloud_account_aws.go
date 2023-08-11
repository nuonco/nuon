package app

type AWSSettings struct {
	Model

	InstallID string

	Region     string `faker:"-"`
	IamRoleArn string
	AccountID  string
}
