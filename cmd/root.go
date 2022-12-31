package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"time"
)

// NewRootCommand creates the root command
func NewRootCommand() (c *cobra.Command) {
	opt := &option{}
	c = &cobra.Command{
		Use:  "ssh-hosts",
		RunE: opt.runE,
	}

	flags := c.Flags()
	flags.StringVarP(&opt.sshDir, "ssh-dir", "", "$HOME/.ssh", "The directory of ssh")
	flags.StringVarP(&opt.etcDir, "etc-dir", "", "/etc", "The directory of etc")
	return
}

func (o *option) runE(c *cobra.Command, args []string) (err error) {
	for {
		select {
		case <-time.After(time.Second * 2):
			if err = o.copyHostRecords(c.Context()); err != nil {
				return
			}
		case <-c.Context().Done():
			return
		}
	}
	return
}

func (o *option) copyHostRecords(c context.Context) (err error) {
	_, cancel := context.WithCancel(c)
	defer cancel()

	var hosts map[string]string
	hosts, err = getHostRecords(o.sshDir)
	if err == nil && len(hosts) > 0 {
		err = writeToHosts(path.Join(o.etcDir, "hosts"), hosts)
	}
	return
}

func getHostRecords(dir string) (records map[string]string, err error) {
	var data []byte
	configPath := os.ExpandEnv(path.Join(dir, "config"))
	if data, err = os.ReadFile(configPath); err != nil {
		return
	}

	records = map[string]string{}
	lines := strings.Split(string(data), "\n")
	var (
		host     string
		hostName string
	)
	for _, line := range lines {
		if strings.HasPrefix(line, "Host ") {
			host = strings.TrimPrefix(line, "Host")
			host = strings.TrimSpace(host)
		}
		if strings.HasPrefix(line, "  HostName ") {
			hostName = strings.TrimPrefix(line, "  HostName ")
			hostName = strings.TrimSpace(hostName)
			records[hostName] = host
		}
	}
	return
}

const (
	beginLine = "# start with ssh-hosts"
	endLine   = "# end with ssh-hosts"
)

func writeToHosts(hostsPath string, hosts map[string]string) (err error) {
	var data []byte
	if data, err = os.ReadFile(hostsPath); err != nil {
		return
	}

	index := -1
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, beginLine) {
			index = i
			break
		}
	}

	existHosts := map[string]string{}
	if index == -1 {
		lines = append(lines, []string{beginLine, endLine}...)
		index = len(lines) - 2
	} else {
		var endIndex int
		existHosts, endIndex = findExistRecords(lines, index)
		lines = append(lines[0:index+1], lines[endIndex:]...)
	}

	// merge records
	for host, hostName := range hosts {
		existHosts[host] = hostName
	}
	hosts = existHosts

	// insert the new records
	for host, hostName := range hosts {
		right := []string{fmt.Sprintf("%s %s", host, hostName)}
		right = append(right, lines[index+1:]...)

		lines = append(lines[0:index+1], right...)
	}
	err = os.WriteFile(hostsPath, []byte(strings.Join(lines, "\n")), 0622)
	return
}

func findExistRecords(lines []string, index int) (hosts map[string]string, endIndex int) {
	hosts = map[string]string{}
	lines = lines[index:]
	for i, line := range lines {
		if strings.HasPrefix(line, endLine) {
			endIndex = i + index
			break
		}

		pair := strings.Split(line, " ")
		if len(pair) == 2 {
			hosts[pair[0]] = pair[1]
		}
	}
	return
}

type option struct {
	sshDir string
	etcDir string
}
