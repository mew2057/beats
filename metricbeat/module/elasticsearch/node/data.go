package node

import (
	"encoding/json"

	"github.com/elastic/beats/libbeat/common"
	s "github.com/elastic/beats/libbeat/common/schema"
	c "github.com/elastic/beats/libbeat/common/schema/mapstriface"
	"github.com/elastic/beats/metricbeat/mb"
)

var (
	schema = s.Schema{
		"name":    c.Str("name"),
		"version": c.Str("version"),
		"jvm": c.Dict("jvm", s.Schema{
			"version": c.Str("version"),
			"memory": c.Dict("mem", s.Schema{
				"heap": s.Object{
					"init": s.Object{
						"bytes": c.Int("heap_init_in_bytes"),
					},
					"max": s.Object{
						"bytes": c.Int("heap_max_in_bytes"),
					},
				},
				"nonheap": s.Object{
					"init": s.Object{
						"bytes": c.Int("non_heap_init_in_bytes"),
					},
					"max": s.Object{
						"bytes": c.Int("non_heap_max_in_bytes"),
					},
				},
			}),
		}),
		"process": c.Dict("process", s.Schema{
			"mlockall": c.Bool("mlockall"),
		}),
	}
)

func eventsMapping(r mb.ReporterV2, content []byte) []error {
	nodesStruct := struct {
		ClusterName string                            `json:"cluster_name"`
		Nodes       map[string]map[string]interface{} `json:"nodes"`
	}{}

	err := json.Unmarshal(content, &nodesStruct)
	if err != nil {
		r.Error(err)
		return []error{err}
	}

	var errs []error

	for name, node := range nodesStruct.Nodes {

		event := mb.Event{}

		event.MetricSetFields, err = eventMapping(node)
		if err != nil {
			errs = append(errs, err)
		}

		// Write name here as full name only available as key
		event.MetricSetFields["name"] = name

		event.ModuleFields = common.MapStr{}
		event.ModuleFields.Put("cluster.name", nodesStruct.ClusterName)

		event.RootFields = common.MapStr{}
		event.RootFields.Put("service.name", "elasticsearch")

		r.Event(event)
	}

	return errs
}

func eventMapping(node map[string]interface{}) (common.MapStr, *s.Errors) {
	return schema.Apply(node)
}
