package cmd

import (
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/ArvinLim/cobra-cli/tpl"
	"github.com/spf13/cobra"
)

// Project contains name, license and paths to projects.
type Project struct {
	// v2
	PkgName      string
	Copyright    string
	AbsolutePath string
	Legal        License
	Viper        bool
	AppName      string
	PkgCmd       string
}

type Command struct {
	CmdName   string
	CmdParent string
	*Project
}

func (p *Project) GetCmdPath() string {
	if p.PkgCmd == "main" {
		return ""
	}

	return p.PkgCmd
}

func (p *Project) Create() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(p.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.Mkdir(p.AbsolutePath, 0754); err != nil {
			return err
		}
	}

	// create main.go
	mainFile, err := os.Create(fmt.Sprintf("%s/main.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer mainFile.Close()

	mainTemplate := template.Must(template.New("main").Parse(string(tpl.MainTemplate())))
	err = mainTemplate.Execute(mainFile, p)
	if err != nil {
		return err
	}

	// create cmd/root.go
	if _, err = os.Stat(fmt.Sprintf("%s/%s", p.AbsolutePath, p.GetCmdPath())); os.IsNotExist(err) {
		cobra.CheckErr(os.Mkdir(fmt.Sprintf("%s/%s", p.AbsolutePath, p.GetCmdPath()), 0751))
	}
	rootFile, err := os.Create(fmt.Sprintf("%s/%s/root.go", p.AbsolutePath, p.GetCmdPath()))
	if err != nil {
		return err
	}
	defer rootFile.Close()

	rootTemplate := template.Must(template.New("root").Parse(string(tpl.RootTemplate())))
	err = rootTemplate.Execute(rootFile, p)
	if err != nil {
		return err
	}

	// create license
	return p.createLicenseFile()
}

func (p *Project) createLicenseFile() error {
	data := map[string]interface{}{
		"copyright": copyrightLine(),
	}
	licenseFile, err := os.Create(fmt.Sprintf("%s/LICENSE", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer licenseFile.Close()

	licenseTemplate := template.Must(template.New("license").Parse(p.Legal.Text))
	return licenseTemplate.Execute(licenseFile, data)
}

func (c *Command) Create() error {
	if c.Project == nil {
		return errors.New("Project empty.")
	}

	p := c.Project
	if _, err := os.Stat(fmt.Sprintf("%s/%s", p.AbsolutePath, p.GetCmdPath())); os.IsNotExist(err) {
		cobra.CheckErr(os.Mkdir(fmt.Sprintf("%s/%s", p.AbsolutePath, p.GetCmdPath()), 0751))
	}

	cmdFile, err := os.Create(fmt.Sprintf("%s/%s/%s.go", c.AbsolutePath, p.GetCmdPath(), c.CmdName))
	if err != nil {
		return err
	}
	defer cmdFile.Close()

	commandTemplate := template.Must(template.New("sub").Parse(string(tpl.AddCommandTemplate())))
	err = commandTemplate.Execute(cmdFile, c)
	if err != nil {
		return err
	}
	return nil
}
