package main

import (
	"errors"
	"log"
	"signachurn/db"
	"signachurn/scan"
	"signachurn/scan/cpp"
	"signachurn/scan/proto"
	"time"
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
	log.Println("starting DB")
	db := db.NewDB()
	db.Start(func() {

		log.Println("init scanners")

		scanners := NewScanners()
		scanners.all = append(scanners.all, &cpp.ScannerCpp{})

		log.Println("adding job")
		job := &proto.ScanJob{
			Type: proto.ScanJobType_SCAN_JOB_GIT,
			Git: &proto.ScanJobGit{
				RemoteURL: "https://github.com/glfw/glfw",
			},
		}

		log.Println("process job")
		scanRes, err := scanners.Process(job)
		if err != nil {
			panic(err)
		}
		log.Println("Found", len(scanRes.GetSignatures()), "signatures")
		log.Println("appending")
		db.AddSignatures("head", scanRes.GetSignatures()...)
	})
	defer db.Stop()

	time.Sleep(time.Second * 100)
}
