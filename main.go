package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (s *Scanners) Aquire(job *proto.ScanJob) scan.Scanner {
	for _, sc := range s.all {
		log.Println("Scanner", sc)
		if sc.Accepts(job) {
			return sc
		}
	}
	return nil
}

func (s *Scanners) Process(job *proto.ScanJob) (*proto.ScanResult, error) {
	sc := s.Aquire(job)
	if s == nil {
		str, err := json.Marshal(job)
		return nil, errors.Join(
			fmt.Errorf("no scanner for job", string(str)),
			err,
		)
	}
	return sc.Scan(job)
}

func main() {
	scanners := NewScanners()
	scanners.all = append(scanners.all, &cpp.ScannerCpp{})

	job := &proto.ScanJob{
		Type: proto.ScanJobType_SCAN_JOB_FILE,
		File: &proto.ScanJobFile{
			FileName: "./samples/example.cpp",
		},
	}

	scanRes, err := scanners.Process(job)
	if err != nil {
		panic(err)
	}
	for _, sig := range scanRes.GetSignatures() {
		log.Println(sig.AsString)
	}

}
