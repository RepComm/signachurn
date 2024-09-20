package scan

import "signachurn/scan/proto"

type Scanner interface {
	Accepts(cfg *proto.ScanJob) bool
	Scan(cfg *proto.ScanJob) (*proto.ScanResult, error)
}
