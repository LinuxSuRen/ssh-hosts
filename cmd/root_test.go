package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

func Test_getHostRecords1(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name        string
		args        args
		wantRecords map[string]string
		wantErr     bool
	}{{

		name: "normal",
		args: args{dir: "data"},
		wantRecords: map[string]string{
			"10.121.218.184": "master",
			"10.121.218.185": "node1",
			"10.121.218.186": "node2",
			"10.121.218.242": "vm",
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRecords, err := getHostRecords(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRecords, tt.wantRecords) {
				t.Errorf("getHostRecords() gotRecords = %v, want %v", gotRecords, tt.wantRecords)
			}
		})
	}
}

func Test_writeToHosts(t *testing.T) {
	type args struct {
		dir   string
		hosts map[string]string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantText string
	}{{
		name: "without placeholder",
		args: args{
			dir: "data/hosts",
			hosts: map[string]string{
				"192.168.1.1": "vm",
			},
		},
		wantErr: false,
		wantText: `192.168.1.6     host.docker.internal
# start with ssh-hosts
192.168.1.1 vm
# end with ssh-hosts`,
	}, {
		name: "with placeholder",
		args: args{
			dir: "data/hosts-with-placeholder",
			hosts: map[string]string{
				"192.168.1.1": "vm1",
			},
		},
		wantErr: false,
		wantText: `192.168.1.6     host.docker.internal
# start with ssh-hosts
192.168.1.1 vm1
# end with ssh-hosts`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp(os.TempDir(), "hosts")
			assert.Nil(t, err)
			defer os.RemoveAll(f.Name())
			data, err := os.ReadFile(tt.args.dir)
			assert.Nil(t, err)
			err = os.WriteFile(f.Name(), data, 0622)
			assert.Nil(t, err)

			if err := writeToHosts(f.Name(), tt.args.hosts); (err != nil) != tt.wantErr {
				t.Errorf("writeToHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			data, err = os.ReadFile(f.Name())
			assert.Nil(t, err)
			assert.Equal(t, tt.wantText, string(data))
		})
	}
}

func Test_findExistRecords(t *testing.T) {
	type args struct {
		lines []string
		index int
	}
	tests := []struct {
		name         string
		args         args
		wantHosts    map[string]string
		wantEndIndex int
	}{{
		name: "normal",
		args: args{
			lines: []string{
				"192.168.1.1 fake.com",
				beginLine,
				"140.82.114.3 github.com",
				endLine,
				"192.168.1.1 fake.com",
			},
			index: 1,
		},
		wantHosts: map[string]string{
			"140.82.114.3": "github.com",
		},
		wantEndIndex: 3,
	}, {
		name: "empty record",
		args: args{
			lines: []string{
				"192.168.1.1 fake.com",
				beginLine,
				endLine,
				"192.168.1.1 fake.com",
			},
			index: 1,
		},
		wantHosts:    map[string]string{},
		wantEndIndex: 2,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHosts, gotEndIndex := findExistRecords(tt.args.lines, tt.args.index)
			assert.Equalf(t, tt.wantHosts, gotHosts, "findExistRecords(%v, %v)", tt.args.lines, tt.args.index)
			assert.Equalf(t, tt.wantEndIndex, gotEndIndex, "findExistRecords(%v, %v)", tt.args.lines, tt.args.index)
		})
	}
}
