package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"code.google.com/p/go.crypto/ssh"
	logger "github.com/dictybase/webhooks"
	"github.com/pkg/sftp"

	"gopkg.in/codegangsta/cli.v0"
	"gopkg.in/gin-gonic/gin.v0"
	"gopkg.in/yaml.v1"
)

type Config struct {
	ServiceHost string `yaml:"host,flow"`
	KeyFile     string
	Remotes     []string
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

var log *logger.Logger

func init() {
	log = logger.NewLogger(os.Stderr, logger.INFO)
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
	// check for secret token
	if token := os.Getenv("WEBHOOK_TOKEN"); len(token) == 0 {
		log.Fatal("Environment WEBHOOK_TOKEN is not set")
	}
	config, err := readConfig(c)
	if err != nil {
		log.Fatalf("error in reading config\n%s\n", err)
	}
	handler, err := getHttpHandler(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("starting webserver at %s\n", config.ServiceHost)
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
	// sftp client setup
	clients := getSftpClient(conf)
	// setup routes
	resource := &webHookResource{conf, clients}
	r := gin.Default()
	auth := r.Group("/webhook")
	auth.POST("/email", resource.sendEmail)
	auth.POST("/copy", SecureWebhook(), resource.copyRemote)
	return r, nil
}

func validateToken(messageMAC, body, token []byte) bool {
	mac := hmac.New(sha1.New, token)
	mac.Write(body)
	expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal(messageMAC, []byte(expected))
}

func getSftpClient(conf Config) []*sftp.Client {
	// process the keyfile
	buf, err := ioutil.ReadFile(conf.KeyFile)
	if err != nil {
		log.Fatalf("error in reading private key file %s\n", err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		log.Fatalf("error in parsing private key %s\n", key)
	}
	// client config
	config := &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
	}
	// connection
	clients := make([]*sftp.Client, 0)
	for _, r := range conf.Remotes {
		c, err := ssh.Dial("tcp", r, config)
		if err != nil {
			log.Fatalf("error in ssh connection %s\n", err)
		}
		// sftp handler
		sftp, err := sftp.NewClient(c)
		if err != nil {
			log.Fatalf("error in sftp connection %s\n", err)
		}
		clients = append(clients, sftp)
	}
	return clients
}

func SecureWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validating the hash signature send with the webhook
		// It should match with the given secret token
		messageMAC := c.Request.Header.Get("X-Hub-Signature")
		if len(messageMAC) == 0 {
			log.Error("no digest given in the webhook")
			c.Fail(http.StatusBadRequest, errors.New("no digest given in the webhook"))
			return
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error(err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		//make a copy of request body for subsequent use
		rd := ioutil.NopCloser(bytes.NewBuffer(body))
		if !validateToken([]byte(messageMAC), body, []byte(os.Getenv("WEBHOOK_TOKEN"))) {
			log.Error("unable to validate the hash signature")
			c.Fail(http.StatusBadRequest, errors.New("unable to validate the hash signature"))
			return
		}
		// assign the copy to the request body since the original will get drained by
		// the previous process
		c.Request.Body = rd
		// Validation ends
		c.Next()
	}
}

type webHookResource struct {
	config  Config
	clients []*sftp.Client
}

func (w *webHookResource) Bind(c *gin.Context, s *webHookPush) bool {
	//body, err := ioutil.ReadAll(c.Request.Body)
	//if err != nil {
	//log.Errorf("Unable to read body of requrest error:%s\n", err)
	//c.Fail(http.StatusBadRequest, err)
	//return false
	//}
	d := json.NewDecoder(c.Request.Body)
	err := d.Decode(s)
	if err != nil {
		log.Errorf("Unable to decode json error:%s\n", err)
		c.Fail(http.StatusBadRequest, err)
		return false
	}
	return true
}

func (w *webHookResource) copyRemote(c *gin.Context) {

	var wh webHookPush
	if !w.Bind(c, &wh) {
		log.Error("could not parse json request")
		return
	}
	cdir := filepath.Join(w.config.Workdir, filepath.Base(wh.Repository.CloneUrl))
	g := &gitCmd{
		repository:    wh.Repository.CloneUrl,
		dir:           cdir,
		defaultBranch: wh.Repository.DefaultBranch,
	}
	info, err := os.Stat(cdir)
	if err != nil {
		if os.IsNotExist(err) { // clone repository
			if err := g.Clone(); err != nil {
				log.Error(err)
				c.Fail(http.StatusBadRequest, err)
				return
			}
			log.Infof("cloned repository %s\n", wh.Repository.CloneUrl)

		} else {
			log.Error(err)
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	} else if info.IsDir() { // repository already cloned
		// pull latest changes
		if err := g.Pull(); err != nil {
			log.Error(err)
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		log.Infof("pulled repository %s\n", wh.Repository.CloneUrl)

	} else {
		log.Error("Unknown error, cannot pull or clone repository")
		c.String(http.StatusBadRequest, "Unknown error, cannot pull or clone repository")
		return
	}
	// now checkout particular commit
	if err := g.Checkout(wh.HeadCommit.Id); err != nil {
		log.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Infof("checked out commit %s of repository %s\n", wh.HeadCommit.Id, wh.Repository.CloneUrl)

	// figure out the added/modified file(s) and
	// copy them to remote server
	for _, a := range wh.HeadCommit.Added {
		if err := w.secureCopy(a, cdir); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			log.Error(err)
			return
		}
		log.Infof("copied file %s to all remotes in folder %s", filepath.Join(cdir, a), w.config.Folder)
	}
	for _, m := range wh.HeadCommit.Modified {
		if err := w.secureCopy(m, cdir); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			log.Error(err)
			return
		}
		log.Infof("copied file %s to all remotes in folder %s", filepath.Join(cdir, m), w.config.Folder)
	}
	c.String(200, "copied all files to remote")
	return
}

func (w *webHookResource) secureCopy(file string, dir string) error {
	conf := w.config
	rfile := filepath.Join(conf.Folder, filepath.Base(file))
	for _, cl := range w.clients {
		wr, err := cl.Create(rfile)
		if err != nil {
			return fmt.Errorf("error in creating remote file %s\n", err)
		}
		rd, err := os.Open(filepath.Join(dir, file))
		if err != nil {
			return err
		}
		defer rd.Close()
		if _, err := io.Copy(wr, rd); err != nil {
			return fmt.Errorf("error in copying file to remote system %s\n", err)
		}
	}
	return nil
}

func (w *webHookResource) sendEmail(c *gin.Context) {
}

type gitCmd struct {
	repository    string
	dir           string
	defaultBranch string
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
	cmd := exec.Command("git", "pull", "origin", g.defaultBranch)
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
