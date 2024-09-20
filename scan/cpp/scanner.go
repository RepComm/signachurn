package cpp

import (
	"log"
	"path"
	"signachurn/scan/proto"

	"github.com/go-clang/clang-v14/clang"
)

type ScannerCpp struct {
}

func (s *ScannerCpp) Accepts(cfg *proto.ScanJob) bool {
	if cfg.Type != proto.ScanJobType_SCAN_JOB_FILE {
		return false
	}
	ext := path.Ext(cfg.GetFile().FileName)
	switch ext {
	case ".cpp", ".c", ".hpp", ".h":
		return true
	}
	return true
}
func (s *ScannerCpp) Scan(cfg *proto.ScanJob) (*proto.ScanResult, error) {

	f := cfg.GetFile()

	c := clang.NewIndex(0, 0)
	defer c.Dispose()

	tu := c.ParseTranslationUnit(
		f.FileName, nil,
		nil,
		clang.DefaultEditingTranslationUnitOptions(),
	)
	defer tu.Dispose()

	cursor := tu.TranslationUnitCursor()
	cursor.Visit(func(cursor, parent clang.Cursor) (status clang.ChildVisitResult) {
		if cursor.Kind() == clang.Cursor_FunctionDecl {
			log.Println(
				cursor.Spelling(),
				cursor.Type().Spelling(),
				cursor.ResultType().Spelling(),
			)

		}
		return clang.ChildVisit_Continue
	})

	return nil, nil
}
