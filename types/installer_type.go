package types

type ContextType string

const (
	Trigger ContextType = "trigger_type"
)

type TriggerType string

const (
	HTTP_CHECK   TriggerType = "http-request-check"
	HTTP_REQUEST TriggerType = "http-request-trigger"
	CRON_JOB     TriggerType = "cron-job-trigger"
	INSTALL      TriggerType = "install-trigger"
)

type STATUS_MSG string

const (
	// 1. FetchUpdate
	OUT_OF_DATE     STATUS_MSG = "out-of-date"
	READY_TO_UPDATE STATUS_MSG = "ready-to-update"
	UP_TO_DATE      STATUS_MSG = "up-to-date"

	// 2. Install
	FETCHING    = "fetching"
	DOWNLOADING = "downloading"
	DECOMPRESS  = "decompressing"
	INSTALLING  = "installing"
	RESTARTING  = "restarting"
	MIGRATION   = "migration"
	OTHER       = "other"
)
