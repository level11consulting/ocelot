package credsadd

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/deserialize"
	"github.com/level11consulting/orbitalci/client/commandhelper"
	models "github.com/level11consulting/orbitalci/models/pb"
	"io/ioutil"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI      cli.Ui
	flags   *flag.FlagSet
	fileloc string
	config  *commandhelper.ClientConfig
}

func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.config.Client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *commandhelper.ClientConfig {
	return c.config
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing creds to upload")
}

func (c *cmd) runCredFileUpload(ctx context.Context) int {
	credWrap := &models.AllCredsWrapper{}
	dese := deserialize.New()
	confFile, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	if err = dese.YAMLToStruct(confFile, credWrap); err != nil {
		c.UI.Error(fmt.Sprintf("Could not process file, please check documentation\nError: %s", err.Error()))
		return 1
	}
	var errOccured bool
	if len(credWrap.VcsCreds.Vcs) == 0 {
		c.UI.Error("Did not read any credentials! Is your yaml formatted correctly?")
		return 1
	}

	if credWrap.VcsCreds != nil {
		for _, configVal := range credWrap.VcsCreds.Vcs {
			_, err = c.config.Client.SetVCSCreds(ctx, configVal)
			if err != nil {
				c.UI.Error(fmt.Sprintf("Could not add vcs credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
				errOccured = true
			} else {
				c.UI.Info(fmt.Sprintf("Added vcs credentials for account: %s", configVal.AcctName))

				//after creds are successfully uploaded via file, upload ssh key file accordingly
				if len(configVal.SshFileLoc) > 0 {
					c.UI.Info(fmt.Sprintf("\tdetected ssh file location: %s", configVal.SshFileLoc))
					commandhelper.UploadSSHKeyFile(ctx, c.UI, c.config.Client, configVal.AcctName, configVal.SubType, configVal.SshFileLoc)
				}
			}
		}
	}

	if credWrap.RepoCreds != nil {
		for _, configVal := range credWrap.RepoCreds.Repo {
			_, err = c.config.Client.SetRepoCreds(ctx, configVal)
			if err != nil {
				c.UI.Error(fmt.Sprintf("Could not add repo credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
				errOccured = true
			} else {
				c.UI.Info(fmt.Sprintf("Added repo credentials for account: %s", configVal.AcctName))
			}
		}
	}

	if errOccured {
		return 1
	}
	return 0
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	if c.fileloc != "" {
		return c.runCredFileUpload(ctx)
	} else {
		c.UI.Error("credfile-loc required, see help")
		return 1
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

//TODO: creds repo takes in different file than creds command??? it makes no sense
const synopsis = "Add credentials via yaml file"
const help = `
Usage: ocelot creds add --credfile-loc ~/credfile-yaml.yaml
  Add one set of credentials or a list of them using a yaml file specified by --credfile-loc <yaml_file>
  This client endpoint accepts both vcs and repo, and must be in the following format:

	vcsCreds:
	  vcs:
	  - clientId: fancy-frickin-identification
		clientSecret: SHH-BE-QUIET-ITS-A-SECRET
		tokenURL: https://ocelot.perf/site/oauth2/access_token
		acctName: lamb-shank
		type: bitbucket
		sshFileLoc: path_to_your_ssh_key
	repoCreds:
	  repo:
	  - username: thisBeMyUserName
		password: SHH-BE-QUIET-ITS-A-SECRET
		repoUrl: https://ocelot.perf/nexus-yo
		acctName: jessishank
		type: nexus

  sshFileLoc is an optional field to set when uploading your vcs credentials, if added, it will be used 
  for cloning repositories belonging to that account
`
