package label 

import (
	"github.com/grammeaway/lazyhetzner/internal/resource"
)

type LabelsLoadedMsg struct {
	Labels map[string]string
	RelatedResourceType resource.ResourceType
	RelatedResourceName string
}


