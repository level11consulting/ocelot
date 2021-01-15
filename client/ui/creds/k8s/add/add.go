package kubeadd

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/level11consulting/orbitalci/client/commandhelper"
	models "github.com/level11consulting/orbitalci/models/pb"
	"io/ioutil"
	"strings"
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
	account string
	name    string
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

	c.flags.StringVar(&c.fileloc, "kubeconfig", "ERROR",
		"Location of kubeconfig file to upload")
	c.flags.StringVar(&c.account, "acct", "ERROR",
		"Account name to file kubeconfig under")
	c.flags.StringVar(&c.name, "name", "ERROR",
		"Name for kubeconfig (Using your cluster name is recommended)")
}

// uploadCredential will check if credential already exists. if it does, it will ask if the user wishes to overwrite. if the user responds YES, the credential will be updated.
// if it does not exist, will be inserted as normal.
func uploadCredential(ctx context.Context, client models.GuideOcelotClient, UI cli.Ui, cred *models.K8SCreds) error {
	exists, err := client.K8SCredExists(ctx, cred)
	if err != nil {
		return err
	}

	if exists.Exists {
		update, err := UI.Ask(fmt.Sprintf("Account Name %s already has a Repo Type %s config with the name %s. Do you want to overwrite? "+
			"Only a YES will continue with update, otherwise the client will exit. ", cred.AcctName, strings.ToLower(cred.SubType.String()), cred.Identifier))
		if err != nil {
			return err
		}
		if update != "YES" {
			UI.Info("Did not recieve a YES at the prompt, will not overwrite. Exiting.")
			return &commandhelper.DontOverwrite{}
		}
		_, err = client.UpdateK8SCreds(ctx, cred)
		if err != nil {
			return err
		}
		UI.Error("Succesfully update K8s Credential.")
		return nil
	}
	_, err = client.SetK8SCreds(ctx, cred)
	return err
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	k8cred := &models.K8SCreds{SubType: models.SubCredType_KUBECONF}
	if c.account == "ERROR" {
		c.UI.Error("-acct was not provided")
		return 1
	}
	k8cred.AcctName = c.account
	if c.fileloc == "ERROR" {
		c.UI.Error("-kubeconfig required")
		return 1
	}
	k8cred.Identifier = c.name
	if c.name == "ERROR" {
		c.UI.Error("-name required")
		return 1
	}
	kubeconf, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	k8cred.K8SContents = string(kubeconf)

	if err = uploadCredential(ctx, c.config.Client, c.UI, k8cred); err != nil {
		if _, ok := err.(*commandhelper.DontOverwrite); ok {
			return 0
		}
		c.UI.Error("Could not add Kubernetes kubeconfig to admin")
		commandhelper.UIErrFromGrpc(err, c.UI, err.Error())
		return 1
	}
	c.UI.Info("Successfully added a kubeconfig to the account " + c.account)
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Add a kubeconfig for connection with kubernetes to ocelot"
const help = `
Usage: ocelot creds k8s add -acct my_kewl_acct -name cluster_name -kubeconfig=/home/user/.kube/cluster-config.yaml

`
