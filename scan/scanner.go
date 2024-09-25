package scan

import "signachurn/scan/proto"

/**Suggests trying another scanner, couldn't find content to scan matching our scanner's critieria*/
type NoValidContentError struct {
	Err error
}

func (e *NoValidContentError) Error() string {
	return e.Err.Error()
}

type NoScannerForJobError struct {
	Err error
}

func (e *NoScannerForJobError) Error() string {
	return e.Err.Error()
}

/**Refusing to process job, usually scanner.Accept() returns false first, but this is fired when the scanner thinks it's being provided the job anyways*/
type UnacceptableJobError struct {
	Err error
}

func (e *UnacceptableJobError) Error() string {
	return e.Err.Error()
}

type Scanner interface {
	Accepts(cfg *proto.ScanJob) bool
	Scan(cfg *proto.ScanJob) ([]*proto.ScanResult, error)
}

const CLONES_DIR string = "./clones"
