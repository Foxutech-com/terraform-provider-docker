package docker

import (
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/hashicorp/terraform/helper/schema"
)

func flattenServiceMode(in swarm.ServiceMode) []interface{} {
	m := make(map[string]interface{})
	if in.Replicated != nil {
		m["replicated"] = flattenReplicated(in.Replicated)
	}
	if in.Global != nil {
		m["global"] = true
	} else {
		m["global"] = false
	}

	return []interface{}{m}
}

func flattenReplicated(in *swarm.ReplicatedService) []interface{} {
	var out = make([]interface{}, 0, 0)
	m := make(map[string]interface{})
	if in != nil {
		if in.Replicas != nil {
			replicas := int(*in.Replicas)
			m["replicas"] = replicas
		}
	}
	out = append(out, m)
	return out
}

func flattenServiceHosts(in []string) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		split := strings.Split(v, ":")
		m["host"] = split[0]
		m["ip"] = split[1]
		out[i] = m
	}
	return out
}

func flattenServiceNetworks(in []swarm.NetworkAttachmentConfig) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		out[i] = v.Target
	}
	return out
}

func flattenServiceMounts(in []mount.Mount) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["target"] = v.Target
		m["source"] = v.Source
		m["type"] = v.Type
		if len(v.Consistency) > 0 {
			m["consistency"] = v.Consistency
		}
		m["read_only"] = v.ReadOnly
		if v.BindOptions != nil {
			m["bind_propagation"] = v.BindOptions.Propagation
		}
		if v.VolumeOptions != nil {
			m["volume_no_copy"] = v.VolumeOptions.NoCopy
			m["volume_labels"] = v.VolumeOptions.Labels
			if v.VolumeOptions.DriverConfig != nil {
				m["volume_driver_name"] = v.VolumeOptions.DriverConfig.Name
				m["volume_driver_options"] = v.VolumeOptions.DriverConfig.Options
			}
		}
		if v.TmpfsOptions != nil {
			m["tmpfs_size_bytes"] = v.TmpfsOptions.SizeBytes
			m["tmpfs_mode"] = v.TmpfsOptions.Mode.Perm
		}
		out[i] = m
	}
	return out
}

func flattenServiceConfigs(in []*swarm.ConfigReference) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["config_id"] = v.ConfigID
		if len(v.ConfigName) > 0 {
			m["config_name"] = v.ConfigName
		}
		if v.File != nil {
			m["file_name"] = v.File.Name
		}
		out[i] = m
	}
	return out
}

func flattenServiceSecrets(in []*swarm.SecretReference) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["secret_id"] = v.SecretID
		if len(v.SecretName) > 0 {
			m["secret_name"] = v.SecretName
		}
		if v.File != nil {
			m["file_name"] = v.File.Name
		}
		out[i] = m
	}
	return out
}

func flattenServicePorts(in []swarm.PortConfig) []interface{} {
	if in == nil || len(in) == 0 {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["internal"] = int(v.TargetPort)
		if v.PublishedPort > 0 {
			m["external"] = v.PublishedPort
		}
		m["publish_mode"] = v.PublishMode
		m["protocol"] = v.Protocol
		out[i] = m
	}
	return out
}

func flattenServiceUpdateOrRollbackConfig(in *swarm.UpdateConfig) []interface{} {
	var out = make([]interface{}, 0, 0)
	if in == nil {
		return out
	}

	m := make(map[string]interface{})
	m["parallelism"] = in.Parallelism
	m["delay"] = shortDur(in.Delay)
	m["failure_action"] = in.FailureAction
	m["monitor"] = shortDur(in.Monitor)
	m["max_failure_ratio"] = strconv.FormatFloat(float64(in.MaxFailureRatio), 'f', 1, 64)
	m["order"] = in.Order
	out = append(out, m)
	return out
}

func flattenServicePlacement(in *swarm.Placement) []interface{} {
	if in == nil {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, 1, 1)
	m := make(map[string]interface{})
	if len(in.Constraints) > 0 {
		m["constraints"] = newStringSet(schema.HashString, in.Constraints)
	}
	if len(in.Preferences) > 0 {
		m["prefs"] = flattenPlacementPrefs(in.Preferences)
	}
	if len(in.Platforms) > 0 {
		m["platforms"] = flattenPlacementPlatforms(in.Platforms)
	}
	out[0] = m
	return out
}

func flattenPlacementPrefs(in []swarm.PlacementPreference) *schema.Set {
	if in == nil || len(in) == 0 {
		return schema.NewSet(schema.HashString, make([]interface{}, 0, 0))
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		out[i] = v.Spread.SpreadDescriptor
	}
	return schema.NewSet(schema.HashString, out)
}

func flattenPlacementPlatforms(in []swarm.Platform) *schema.Set {
	if in == nil || len(in) == 0 {
		return schema.NewSet(schema.HashString, make([]interface{}, 0, 0))
	}

	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["architecture"] = v.Architecture
		m["os"] = v.OS
		out[i] = m
	}
	return schema.NewSet(schema.HashString, out)
}

func flattenServiceLogging(in *swarm.Driver) []interface{} {
	if in == nil {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, 1, 1)
	m := make(map[string]interface{})
	m["driver_name"] = in.Name
	if len(in.Options) > 0 {
		m["options"] = in.Options
	}
	out[0] = m
	return out
}

func flattenServiceHealthcheck(in *container.HealthConfig) []interface{} {
	if in == nil {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, 1, 1)
	m := make(map[string]interface{})
	if len(in.Test) > 0 {
		m["test"] = in.Test
	}
	m["interval"] = shortDur(in.Interval)
	m["timeout"] = shortDur(in.Timeout)
	m["start_period"] = shortDur(in.StartPeriod)
	m["retries"] = in.Retries
	out[0] = m
	return out
}

func flattenServiceDNSConfig(in *swarm.DNSConfig) []interface{} {
	if in == nil {
		return make([]interface{}, 0, 0)
	}

	var out = make([]interface{}, 1, 1)
	m := make(map[string]interface{})
	if len(in.Nameservers) > 0 {
		m["nameservers"] = in.Nameservers
	}
	if len(in.Search) > 0 {
		m["search"] = in.Search
	}
	if len(in.Options) > 0 {
		m["options"] = in.Options
	}
	out[0] = m
	return out
}

// HELPERS
func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in), len(in))
	for i, v := range in {
		out[i] = v
	}
	return schema.NewSet(f, out)
}
