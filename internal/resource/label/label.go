package label 

import (
	"lazyhetzner/internal/resource"
)

type LabelsLoadedMsg struct {
	Labels map[string]string
	RelatedResourceType resource.ResourceType
	RelatedResourceName string
}


