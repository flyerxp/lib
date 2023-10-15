package logger

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"time"
)

type DataContainer struct {
	NoticeData cmap.ConcurrentMap[string, *noticeData]
	WarnData   cmap.ConcurrentMap[string, *warMetrics]
	ErrData    cmap.ConcurrentMap[string, *errMetrics]
	isInitEnd  bool
	Clear      cmap.ConcurrentMap[string, *time.Timer]
}

var dataContainer *DataContainer
