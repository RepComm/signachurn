package main

import (
	"errors"
	"log"
	"signachurn/scan"
	"signachurn/scan/cpp"
	"signachurn/scan/proto"
)

type Scanners struct {
	all []scan.Scanner
}

func NewScanners() *Scanners {
	result := new(Scanners)
	result.all = make([]scan.Scanner, 0)
	return result
}

func (s *Scanners) Process(job *proto.ScanJob) (*proto.ScanResult, error) {
	for _, sc := range s.all {
		if sc.Accepts(job) {
			res, err := sc.Scan(job)
			if err != nil {
				if _, ok := err.(*scan.NoValidContentError); ok {
					continue
				}
				return res, err
			}
			return res, nil
		}
	}

	return nil, &scan.NoScannerForJobError{
		Err: errors.New("no scanner for job"),
	}
}

func main() {
	scanners := NewScanners()
	scanners.all = append(scanners.all, &cpp.ScannerCpp{})

	job := &proto.ScanJob{
		Type: proto.ScanJobType_SCAN_JOB_GIT,
		Git: &proto.ScanJobGit{
			RemoteURL: "https://github.com/glfw/glfw",
		},
	}

	scanRes, err := scanners.Process(job)
	if err != nil {
		panic(err)
	}
	// for _, sig := range scanRes.GetSignatures() {
	// 	log.Println("Signature", sig.Name, sig.AsString)
	// }
	log.Println("Found", len(scanRes.GetSignatures()), "signatures")

}
