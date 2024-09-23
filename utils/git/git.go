package git

import (
	"errors"
	"log"
	"os/exec"
	"path"
	"signachurn/utils/files"
	"strings"
)

func ListTags() ([]string, error) {
	c := exec.Command(
		"git",
		"--no-pager",
		"tag",
		"-l",
	)
	o, err := c.Output()
	if err != nil {
		return nil, errors.Join(errors.New("couldnt list git tags"), err)
	}
	return strings.Split(string(o), "\n"), nil
}

func CloneRemote(url string, DirPath string) error {
	log.Println("cloning remote", url)
	c := exec.Command("git", "clone", url, DirPath)
	// c.Dir = DirPath
	err := c.Run()
	if err != nil {
		return errors.Join(errors.New("couldnt clone remote repo"), err)
	}
	return nil
}

func Checkout(tag string) error {
	log.Println("git: checkout", tag)
	c := exec.Command("git", "checkout", tag)
	err := c.Run()
	if err != nil {
		return errors.Join(errors.New("couldnt checkout tag"), err)
	}
	return nil
}

func GetRemoteURL(DirPath string) (string, error) {
	c := exec.Command(
		"git",
		"config",
		"--get",
		"remote.origin.url",
	)
	c.Dir = DirPath
	o, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(o)), nil
}

func FetchRemoteLatest(DirPath string) error {
	log.Println("git: fetch remote latest", DirPath)
	c := exec.Command(
		"git",
		"fetch",
		"origin",
	)
	c.Dir = DirPath
	err := c.Run()
	if err != nil {
		return errors.Join(
			errors.New("couldnt fetch origin"),
			err,
		)
	}
	c = exec.Command(
		"git",
		"reset",
		"--hard",
		"origin",
	)
	c.Dir = DirPath
	err = c.Run()
	if err != nil {
		return errors.Join(
			errors.New("couldnt reset local to origin head"),
			err,
		)
	}
	return nil
}

func UrlToDirName(url string) string {
	return path.Base(url)
}

/**If path.Base(url) exists, checks if is remote by URL and attempts updating it
 * if not exists, attempts to clone url
 * if exist check fails, returns os.Stat failure
 */
func EnsureCloned(url string, DestDirPath string) error {
	log.Println("git: ensure cloned", url)

	base := UrlToDirName(url)
	if DestDirPath == "" {
		DestDirPath = path.Join(".", base)
	}

	exists, err := files.FileExists(DestDirPath)
	if err != nil {
		return errors.Join(errors.New("couldnt check if folder exists before git cloning to it"), err)
	}
	//folder already exists, we've probably see repo before
	//or has similar repo name to another
	//we shouldn't process more than one at a time
	if exists {
		//Attempt to get git remote URL from folder
		rurl, err := GetRemoteURL(DestDirPath)
		if err != nil {
			return errors.Join(
				errors.New("couldnt check if folder is a git repo from our remote URL"),
				err,
			)
		}
		//if URL is same as ours, make sure its updated
		if strings.ToLower(rurl) == strings.ToLower(url) {
			err = FetchRemoteLatest(DestDirPath)
			if err != nil {
				return errors.Join(
					errors.New("couldnt fetch latest remote"),
					err,
				)
			}
		}
	} else {
		//no cloned repo detected, clone it
		err := CloneRemote(url, DestDirPath)
		if err != nil {
			return errors.Join(errors.New("folder name does not conflict, but couldnt clone remote"))
		}
	}
	return nil
}
