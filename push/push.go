package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/codegangsta/cli.v0"
	"gopkg.in/gin-gonic/gin.v0"
	"gopkg.in/yaml.v1"
)

type Config struct {
	ServiceHost string `yaml:"host,flow"`
	KeyFile     string
	Remote      string
	User        string
	Folder      string
	Workdir     string
}

type webHookPush struct {
	After   string `json:"after"`
	Before  string `json:"before"`
	Commits []struct {
		Added  []interface{} `json:"added"`
		Author struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"committer"`
		Distinct  bool          `json:"distinct"`
		Id        string        `json:"id"`
		Message   string        `json:"message"`
		Modified  []string      `json:"modified"`
		Removed   []interface{} `json:"removed"`
		Timestamp string        `json:"timestamp"`
		Url       string        `json:"url"`
	} `json:"commits"`
	Compare    string `json:"compare"`
	Created    bool   `json:"created"`
	Deleted    bool   `json:"deleted"`
	Forced     bool   `json:"forced"`
	HeadCommit struct {
		Added  []string `json:"added"`
		Author struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"committer"`
		Distinct  bool     `json:"distinct"`
		Id        string   `json:"id"`
		Message   string   `json:"message"`
		Modified  []string `json:"modified"`
		Removed   []string `json:"removed"`
		Timestamp string   `json:"timestamp"`
		Url       string   `json:"url"`
	} `json:"head_commit"`
	Pusher struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"pusher"`
	Ref        string `json:"ref"`
	Repository struct {
		ArchiveUrl       string      `json:"archive_url"`
		AssigneesUrl     string      `json:"assignees_url"`
		BlobsUrl         string      `json:"blobs_url"`
		BranchesUrl      string      `json:"branches_url"`
		CloneUrl         string      `json:"clone_url"`
		CollaboratorsUrl string      `json:"collaborators_url"`
		CommentsUrl      string      `json:"comments_url"`
		CommitsUrl       string      `json:"commits_url"`
		CompareUrl       string      `json:"compare_url"`
		ContentsUrl      string      `json:"contents_url"`
		ContributorsUrl  string      `json:"contributors_url"`
		CreatedAt        int64       `json:"created_at"`
		DefaultBranch    string      `json:"default_branch"`
		Description      string      `json:"description"`
		DownloadsUrl     string      `json:"downloads_url"`
		EventsUrl        string      `json:"events_url"`
		Fork             bool        `json:"fork"`
		Forks            int64       `json:"forks"`
		ForksCount       int64       `json:"forks_count"`
		ForksUrl         string      `json:"forks_url"`
		FullName         string      `json:"full_name"`
		GitCommitsUrl    string      `json:"git_commits_url"`
		GitRefsUrl       string      `json:"git_refs_url"`
		GitTagsUrl       string      `json:"git_tags_url"`
		GitUrl           string      `json:"git_url"`
		HasDownloads     bool        `json:"has_downloads"`
		HasIssues        bool        `json:"has_issues"`
		HasWiki          bool        `json:"has_wiki"`
		Homepage         interface{} `json:"homepage"`
		HooksUrl         string      `json:"hooks_url"`
		HtmlUrl          string      `json:"html_url"`
		Id               int64       `json:"id"`
		IssueCommentUrl  string      `json:"issue_comment_url"`
		IssueEventsUrl   string      `json:"issue_events_url"`
		IssuesUrl        string      `json:"issues_url"`
		KeysUrl          string      `json:"keys_url"`
		LabelsUrl        string      `json:"labels_url"`
		Language         interface{} `json:"language"`
		LanguagesUrl     string      `json:"languages_url"`
		MasterBranch     string      `json:"master_branch"`
		MergesUrl        string      `json:"merges_url"`
		MilestonesUrl    string      `json:"milestones_url"`
		MirrorUrl        interface{} `json:"mirror_url"`
		Name             string      `json:"name"`
		NotificationsUrl string      `json:"notifications_url"`
		OpenIssues       int64       `json:"open_issues"`
		OpenIssuesCount  int64       `json:"open_issues_count"`
		Owner            struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"owner"`
		Private         bool   `json:"private"`
		PullsUrl        string `json:"pulls_url"`
		PushedAt        int64  `json:"pushed_at"`
		ReleasesUrl     string `json:"releases_url"`
		Size            int64  `json:"size"`
		SshUrl          string `json:"ssh_url"`
		Stargazers      int64  `json:"stargazers"`
		StargazersCount int64  `json:"stargazers_count"`
		StargazersUrl   string `json:"stargazers_url"`
		StatusesUrl     string `json:"statuses_url"`
		SubscribersUrl  string `json:"subscribers_url"`
		SubscriptionUrl string `json:"subscription_url"`
		SvnUrl          string `json:"svn_url"`
		TagsUrl         string `json:"tags_url"`
		TeamsUrl        string `json:"teams_url"`
		TreesUrl        string `json:"trees_url"`
		UpdatedAt       string `json:"updated_at"`
		Url             string `json:"url"`
		Watchers        int64  `json:"watchers"`
		WatchersCount   int64  `json:"watchers_count"`
	} `json:"repository"`
}

func main() {
	app := cli.NewApp()
	app.Name = "webhook"
	app.Usage = "webhook server to communicate with github events"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Name of the config file",
		},
	}
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	config, err := readConfig(c)
	if err != nil {
		log.Fatalf("error in reading config\n%s\n", err)
	}
	handler, err := getHttpHandler(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting webserver at %s\n", config.ServiceHost)
	http.ListenAndServe(config.ServiceHost, handler)
}

func readConfig(c *cli.Context) (Config, error) {
	config := Config{}
	if c.Generic("config") == nil {
		return config, fmt.Errorf("config argument is not given")
	}
	ypath := c.String("config")
	if _, err := os.Stat(ypath); err != nil {
		return config, fmt.Errorf("config file path is not valid")
	}
	ydata, err := ioutil.ReadFile(ypath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal([]byte(ydata), &config)
	return config, err
}

func getHttpHandler(conf Config) (http.Handler, error) {
	// setup routes
	resource := &webHookResource{conf}
	r := gin.Default()
	auth := r.Group("/webhook")
	auth.POST("/email", resource.sendEmail)
	auth.POST("/copy", resource.copyRemote)
	return r, nil
}

type webHookResource struct {
	config Config
}

func (w *webHookResource) copyRemote(c *gin.Context) {
	var wh webHookPush
	if !c.Bind(&wh) {
		log.Println("could not parse json request")
		c.String(400, "could not parse json request")
		return
	}
	cdir := filepath.Join(w.config.Workdir, filepath.Base(wh.Repository.CloneUrl))
	g := &gitCmd{
		repository: wh.Repository.CloneUrl,
		dir:        cdir,
	}
	_, err := os.Stat(cdir)
	// repository already cloned
	if os.IsExist(err) {
		// pull latest changes
		if err := g.Pull(); err != nil {
			log.Println(err)
			c.String(400, err.Error())
			return
		}
		log.Printf("pulled repository %s\n", wh.Repository.CloneUrl)

	} else {
		if err := g.Clone(); err != nil {
			log.Println(err)
			c.String(400, err.Error())
			return
		}
		log.Printf("cloned repository %s\n", wh.Repository.CloneUrl)

	}
	// now checkout particular commit
	if err := g.Checkout(wh.HeadCommit.Id); err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	log.Printf("checked out commit %s of repository %s\n", wh.HeadCommit.Id, wh.Repository.CloneUrl)
	// figure out the modified file(s)
	reportChangedFiles(wh.HeadCommit.Added, cdir, "added", c)
	reportChangedFiles(wh.HeadCommit.Modified, cdir, "modified", c)
	return
	// copy them to remote server
}

func reportChangedFiles(files []string, cdir string, event string, c *gin.Context) {
	for _, a := range files {
		p := filepath.Join(cdir, a)
		log.Printf("%s file %s\n", event, p)
		c.String(200, "%s file %s\n", event, p)
	}
}

func (w *webHookResource) sendEmail(c *gin.Context) {
}

type gitCmd struct {
	repository string
	dir        string
}

func (g *gitCmd) Checkout(b string) error {
	err := os.Chdir(g.dir)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "checkout", "-f", b)
	var eout bytes.Buffer
	cmd.Stderr = &eout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, eout.String())
	}
	return nil
}

func (g *gitCmd) Pull() error {
	err := os.Chdir(g.dir)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "pull", "origin")
	var eout bytes.Buffer
	cmd.Stderr = &eout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, eout.String())
	}
	return nil
}

func (g *gitCmd) Clone() error {
	cmd := exec.Command("git", "clone", g.repository, g.dir)
	var eout bytes.Buffer
	cmd.Stderr = &eout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, eout.String())
	}
	return nil
}
