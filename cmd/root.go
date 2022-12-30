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
	flags.StringVarP(&opt.sshDir, "ssh-dir", "", "~/.ssh", "The directory of ssh")
	flags.StringVarP(&opt.etcDir, "etc-dir", "", "/etc", "The directory of etc")
	return
}

func (o *option) runE(c *cobra.Command, args []string) (err error) {
	for {
		select {
		case <-time.After(time.Second * 2):
			_ = o.copyHostRecords(c.Context())
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
	fmt.Println(hosts)
	if err == nil && len(hosts) > 0 {
		err = writeToHosts(path.Join(o.etcDir, "hosts"), hosts)
	}
	return
}

func getHostRecords(dir string) (records map[string]string, err error) {
	var data []byte
	if data, err = os.ReadFile(path.Join(dir, "config")); err != nil {
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
			records[host] = hostName
		}
	}
	return
}

func writeToHosts(hostsPath string, hosts map[string]string) (err error) {
	var data []byte
	if data, err = os.ReadFile(hostsPath); err != nil {
		return
	}

	index := -1
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# start with ssh-hosts") {
			index = i
			break
		}
	}

	if index == -1 {
		index = len(lines) - 1
		hosts["#"] = "start with ssh-hosts"
	}

	for host, hostName := range hosts {
		var left []string
		if index < len(lines) {
			left = lines[index+1:]
		}
		left = append(left, fmt.Sprintf("%s %s", hostName, host))
		lines = append(lines[0:index+1], left...)
	}
	err = os.WriteFile(hostsPath, []byte(strings.Join(lines, "\n")), 0622)
	return
}

type option struct {
	sshDir string
	etcDir string
}
