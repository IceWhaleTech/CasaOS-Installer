package types

type ContextType string

const (
	Trigger ContextType = "trigger_type"
)

type TriggerType string

// TODO: 考虑重构这里
// 这里是设计是因为最早业务是前端不同方式调用下，状态要有不同的变化，所以区分不同的样式然后来做不同的处理。
// 但是随着 @ETWang1991 的变化，现在不同的调用方式，前端都看不见，考虑删除这里来减少复杂度。
const (
	HTTP_REQUEST TriggerType = "http-request-trigger"
	CRON_JOB     TriggerType = "cron-job-trigger"
	INSTALL      TriggerType = "install-trigger"
)

type STATUS_MSG string

const (
	// 1. FetchUpdate
	OUT_OF_DATE     = "out-of-date"
	READY_TO_UPDATE = "ready-to-update"
	UP_TO_DATE      = "up-to-date"

	// 2. Install
	FETCHING    = "fetching"
	DOWNLOADING = "downloading"
	DECOMPRESS  = "decompressing"
	INSTALLING  = "installing"
	RESTARTING  = "restarting"
	MIGRATION   = "migration"
	OTHER       = "other"
)
