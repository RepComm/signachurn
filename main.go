package main

import (
	"depshit/cmd"
	"depshit/db"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

type TagCommit struct {
	Short string
	Id string
}

func RemoteTags(url string) ([]TagCommit, error) {
	r := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	log.Println("fetching tags")
	refs, err := r.List(&git.ListOptions{
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return nil, err
	}
	results := make([]TagCommit, 0)
	
	for _, ref := range refs {
		name := ref.Name()
		if name.IsTag() {
			tc := TagCommit{
				Short: name.Short(),
				Id: name.String(),
			}
			results = append(results, tc)
		}
	}
	
	return results, nil
}

func main() {
	db,err := db.ConnectDB("file:./depshit.db")
	if err != nil {
		panic(err)
	}

	cmds := cmd.StdLineCmds{}

	shouldClose := false

	exit := func () {
		shouldClose = true
		err := db.Close()
		if err != nil {
			fmt.Println(err)
		}
		cmds.Stop()

		os.Exit(0)
	}

	
	cmds.Start(func(cmd cmd.StdLineCmd) {
		switch cmd.Cmd {
		case "exit":
			fmt.Println("exiting")
			exit()
		case "repo":
			if cmd.Args["list"] == "true" {
				repoInfos, err := db.ListRepos(10, 0)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("listing repos")
				for _,repoInfo := range repoInfos {
					fmt.Println(repoInfo.Url)
				}
				fmt.Println("done listing repos")
			} else if cmd.Args["add"] == "true" {
				url := cmd.Args["url"]
				if url != "" && strings.HasPrefix(url, "http") {
					tags, err := RemoteTags(url)
					if err != nil {
						fmt.Println("could not connect to remote to retrieve tags", err)
						return
					}
					for _, tag := range tags {
						fmt.Println("tag", tag)
					}

					db.AddRepoByURL(url)


					fmt.Println("added repo by url: ", url)
				} else {
					fmt.Println("invalid url", url)
				}
			}
		}
	})
	
	s := http.Server{
		Addr: "0.0.0.0:8080",
	}

	http.HandleFunc("/htmx/", func(w http.ResponseWriter, r *http.Request) {
		
		result := ""

		

		p := r.URL.Path
		switch p {
		case "/htmx/analyse":
			url := r.PostFormValue("url")
			result = fmt.Sprintf(`<div id="analysis">
			Results %s
			</div>`, url)
		}
		w.Write([]byte(result))
	})

	http.Handle("/", http.FileServer(http.Dir("./web")))

	ServerCertPem := "./server.cert.pem"
	ServerKeyPem := "./server.key.pem"
	err = s.ListenAndServeTLS(ServerCertPem, ServerKeyPem)
	if err != nil {
		fmt.Println("failed SSL, possibly no cert/key, using http instead", err)
		err = s.ListenAndServe()
		if err != nil {
			fmt.Println("failed HTTPS and HTTP, panic", err)
		}
	}

	for !shouldClose {
		time.Sleep(time.Millisecond*50)
	}
}
