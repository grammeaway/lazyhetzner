package resource


import ( 

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	tea "github.com/charmbracelet/bubbletea"
	"context"
	"fmt"
	"strconv"
	"strings"
	"github.com/grammeaway/lazyhetzner/internal/message"
)

type ResourceType int

const (
	ResourceServers ResourceType = iota
	ResourceNetworks
	ResourceLoadBalancers
	ResourceVolumes
)

func GetResourceNameFromType(rt ResourceType) string {
	switch rt {
	case ResourceServers:
		return "Servers"
	case ResourceNetworks:
		return "Networks"
	case ResourceLoadBalancers:
		return "Load Balancers"
	case ResourceVolumes:
		return "Volumes"
	default:
		return "Unknown Resource"
	}
}


func loadResources(client *hcloud.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		servers, err := client.Server.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		networks, err := client.Network.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		loadBalancers, err := client.LoadBalancer.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		volumes, err := client.Volume.All(ctx)
		if err != nil {
			return message.ErrorMsg{err}
		}

		return resourcesLoadedMsg{
			servers:       servers,
			networks:      networks,
			loadBalancers: loadBalancers,
			volumes:       volumes,
		}
	}
}



type resourcesLoadedMsg struct {
	servers       []*hcloud.Server
	networks      []*hcloud.Network
	loadBalancers []*hcloud.LoadBalancer
	volumes       []*hcloud.Volume
}



type ResourceLoadStartMsg struct {
	ResourceType ResourceType
}




func StartResourceLoad(rt ResourceType) tea.Cmd {
	return func() tea.Msg {
		return ResourceLoadStartMsg{ResourceType: rt}
	}
}




func getResourceLabels(client *hcloud.Client, resourceType ResourceType, resourceID int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var labels map[string]string

		switch resourceType {
		case ResourceServers:
			server, _, err := client.Server.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if server == nil {
				return message.ErrorMsg{fmt.Errorf("server with ID %d not found", strconv.Itoa(resourceID))}
			}
			labels = server.Labels

		case ResourceNetworks:
			network, _, err := client.Network.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if network == nil {
				return message.ErrorMsg{fmt.Errorf("network with ID %d not found", resourceID)}
			}
			labels = network.Labels

		case ResourceLoadBalancers:
			lb, _, err := client.LoadBalancer.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if lb == nil {
				return message.ErrorMsg{fmt.Errorf("load balancer with ID %d not found", resourceID)}
			}
			labels = lb.Labels

		case ResourceVolumes:
			volume, _, err := client.Volume.Get(ctx, strconv.Itoa(resourceID))
			if err != nil {
				return message.ErrorMsg{err}
			}
			if volume == nil {
				return message.ErrorMsg{fmt.Errorf("volume with ID %d not found", resourceID)}
			}
			labels = volume.Labels

		default:
			return message.ErrorMsg{fmt.Errorf("unknown resource type: %d", resourceType)}
		}

		labelList := make([]string, 0, len(labels))
		for k, v := range labels {
			labelList = append(labelList, fmt.Sprintf("%s: %s", k, v))
		}

		return message.StatusMsg(fmt.Sprintf("Labels for resource ID %d:\n%s", resourceID, strings.Join(labelList, "\n")))
	}
}
