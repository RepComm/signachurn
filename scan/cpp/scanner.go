package cpp

import (
	"errors"
	"fmt"
	"log"
	"path"
	"signachurn/scan"
	"signachurn/scan/proto"
	"signachurn/utils/files"
	"signachurn/utils/git"
	"strings"

	"github.com/go-clang/clang-v14/clang"
)

type ScannerCpp struct {
}

func (s *ScannerCpp) Accepts(cfg *proto.ScanJob) bool {
	if cfg.Type != proto.ScanJobType_SCAN_JOB_GIT {
		return false
	}
	return true
}
func (s *ScannerCpp) Scan(cfg *proto.ScanJob) ([]*proto.ScanResult, error) {
	//NoValidContentError
	if cfg.Type != proto.ScanJobType_SCAN_JOB_GIT {
		return nil, &scan.UnacceptableJobError{
			Err: errors.New("unacceptable job"),
		}
	}
	g := cfg.GetGit()
	DirName := git.UrlToDirName(g.RemoteURL)
	DirPath := path.Join(scan.CLONES_DIR, DirName)

	err := git.EnsureCloned(g.RemoteURL, DirPath)
	if err != nil {
		return nil, err
	}
	exts := []string{
		// ".c",
		// ".cpp",
		".h",
		".hpp",
	}

	tagNames, err := git.ListTags(DirPath)
	results := []*proto.ScanResult{}
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if len(tagName) < 1 {
			continue
		}
		err := git.Checkout(DirPath, tagName)
		if err != nil {
			log.Println("checkout issue, skipping", err)
			continue
		}

		fpaths := files.FindFileExts(DirPath, true, exts)
		if len(fpaths) < 1 {
			return nil, &scan.NoValidContentError{
				Err: fmt.Errorf("couldnt find files with extensions: %s", exts),
			}
		}

		log.Printf("Found %d headers\n", len(fpaths))

		result := &proto.ScanResult{
			TagName:    tagName,
			Signatures: []*proto.Signature{},
		}
		results = append(results, result)

		c := clang.NewIndex(0, 0)
		defer c.Dispose()

		for _, FilePath := range fpaths {
			sigs := FindSigsInFile(FilePath, c)
			result.Signatures = append(result.Signatures, sigs...)
		}
	}

	return results, nil
}

func FindSigsInFile(FilePath string, c clang.Index) []*proto.Signature {
	result := []*proto.Signature{}

	tu := c.ParseTranslationUnit(
		FilePath, nil,
		nil,
		clang.DefaultEditingTranslationUnitOptions(),
	)
	defer tu.Dispose()

	cursor := tu.TranslationUnitCursor()
	cursor.Visit(func(cursor, parent clang.Cursor) (status clang.ChildVisitResult) {
		if cursor.Kind() == clang.Cursor_FunctionDecl {
			sig := &proto.Signature{
				Name:     cursor.Spelling(),
				AsString: fmt.Sprint(cursor.Type().Spelling(), " ", cursor.ResultType().Spelling()),
			}
			result = append(result, sig)
		}
		return clang.ChildVisit_Continue
	})
	return result
}
