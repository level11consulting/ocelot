package commandhelper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/level11consulting/orbitalci/models/pb"
)

var (
	cmdName       = "git"
	//sshGit        = regexp.MustCompile(`git\@\w+\.\w+\:([^\/.]*)\/([^\..]*)[\.git]?`)
	githubGit     = regexp.MustCompile(`git\@github\.\w+\:([^\/.]*)\/([^\..]*)[\.git]?`)
	bitbucketGit  = regexp.MustCompile(`git\@bitbucket\.\w+\:([^\/.]*)\/([^\..]*)[\.git]?`)
	httpsGithub   = regexp.MustCompile(`https:\/\/github\.com\/([^\/.]*)\/([^\..]*)[\.git]?`)
	httpsBb       = regexp.MustCompile(`https:\/\/\w+\@\w+\.org\/([^\/.]*)\/([^\..]*)[\.git]?`)
	httpsBbNoUser = regexp.MustCompile(`https:\/\/\w+\.org\/([^\/.]*)\/([^\..]*)[\.git]?`)
	//httpsBbNoUserNodotGit = regexp.MustCompile(`https:\/\/\w+\.org\/([^\/.]*)\/([^\..]*[^\s-])`)

	regexmaps = map[pb.SubCredType][]*regexp.Regexp{
		pb.SubCredType_BITBUCKET: {bitbucketGit,httpsBb, httpsBbNoUser},
		pb.SubCredType_GITHUB: {githubGit, httpsGithub},
	}
)

func matchThis(data []byte) (string, pb.SubCredType, error) {
	for vcsType, regexz := range regexmaps {
		for _, regex := range regexz {
			if mtch := regex.FindSubmatch(data); mtch != nil {
				// match should only be 2 matches + the original text....
				if len(mtch) != 3 {
					return "", pb.SubCredType_NIL_SCT, errors.New("unexpected match length " + string(len(mtch)))
				}
				return fmt.Sprintf("%s/%s", string(mtch[1]), string(mtch[2])), vcsType, nil
			}
		}
	}
	return "",  pb.SubCredType_NIL_SCT, errors.New(fmt.Sprintf("did not find an account repo match for the remote origin url %s, please inspect url then contact developers to get new match added.", string(data)))
}

// FindAcctRepo will attempt to run a git command and parse out the acct/repo from it.
func FindAcctRepo() (acctRepo string, credType pb.SubCredType, err error) {
	var cmdOut []byte
	getOrigin := []string{"config", "--get", "remote.origin.url"}
	if cmdOut, err = exec.Command(cmdName, getOrigin...).Output(); err != nil {
		return
	}
	return matchThis(cmdOut)
}

//FindCurrentHash will attempt to grab a hash based on running git commands - see client/output/output.go for usage
func FindCurrentHash() string {
	var (
		cmdOut  []byte
		cmdHash []byte
		err     error
	)

	getBranch := []string{"rev-parse", "--abbrev-ref", "HEAD"}
	if cmdOut, err = exec.Command(cmdName, getBranch...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command to find the current branch: ", err)
	}

	if len(getBranch) > 0 {
		// todo: add origin assumption to docs
		// todo: this fails in a weird way if if the branch hasn't been pushed yet

		remoteBranch := fmt.Sprintf("origin/%s", string(cmdOut))
		if cmdHash, err = exec.Command(cmdName, "rev-parse", strings.TrimSpace(remoteBranch)).Output(); err != nil {
			fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command to find the most recently pushed commit: ", err)
		}
	}

	sha := strings.TrimSpace(string(cmdHash))
	return sha
}
