package nexus

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/integrations"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bytes"
	"errors"
	"fmt"
	"text/template"
)

var settingsXml = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.1.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.1.0 http://maven.apache.org/xsd/settings-1.1.0.xsd">
  <servers>
	{{$username:=.Username}}
	{{$password:=.Password}}
	{{ range $name, $url := .RepoUrl }}
    <server>
      <id>{{$name}}</id>
      <username>{{$username}}</username>
      <password>{{$password}}</password>
    </server>
	{{ end }}
  </servers>
  <profiles>
    <profile>
      <id>level11consulting</id>
      <activation>
        <activeByDefault>true</activeByDefault>
      </activation>
      <repositories>
        <repository>
          <id>ocelotNexus</id>
          <name>Ocelot Rendered</name>
          <url>${env.NEXUS_PUBLIC_M2}</url>
        </repository>
      </repositories>
    </profile>
  </profiles>
</settings>`

//var templ = template.Must(template.New("settingsxml").Parse(settingsXml))

// GetSettingsXml will render and return a maven settings.xml with credentials correlating to the accountName provided
// todo: include project name for further filtering
func GetSettingsXml(rc cred.CVRemoteConfig, accountName string) (string, error) {
	templ, err := template.New("settingsxml").Parse(settingsXml)
	if err != nil {
		return "", err
	}
	repod := models.NewRepoCreds()
	credz, err := rc.GetCredAt(fmt.Sprintf(cred.Nexus, accountName), false, repod)
	if err != nil {
		return "", err
	}
	nexusCred, ok := credz[cred.BuildCredKey("nexus", accountName)]
	if !ok {
		return "", integrations.NCErr("no creds found")
	}
	casted, ok := nexusCred.(*models.RepoCreds)
	if !ok {
		return "", errors.New("unable to cast to RepoCreds, which just shouldn't happen")
	}
	var settings bytes.Buffer
	err = templ.Execute(&settings, casted)
	if err != nil {
		return "", errors.New("unable to render settings.xml template for nexus credentials. error: " + err.Error())
	}
	return settings.String(), nil
}