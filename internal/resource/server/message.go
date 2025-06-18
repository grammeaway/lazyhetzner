package server


import (
	"github.com/hetznercloud/hcloud-go/hcloud"
)


type ServersLoadedMsg struct {
	Servers []*hcloud.Server
}


type ServerSnapshotCreationStartedMsg struct {
	Server *hcloud.Server
	SnapshotName string
}
