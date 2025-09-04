package client

import "time"

// PackageToMap converts Package to map for backward compatibility
func PackageToMap(p *Package) map[string]interface{} {
	return map[string]interface{}{
		"name":        p.Name,
		"description": p.Description,
		"latest":      p.Latest,
		"versions":    p.Versions,
		"tags":        p.Tags,
		"updated_at":  p.UpdatedAt,
	}
}

// PackageVersionToMap converts PackageVersion to map
func PackageVersionToMap(pv *PackageVersion) map[string]interface{} {
	return map[string]interface{}{
		"name":         pv.Name,
		"version":      pv.Version,
		"description":  pv.Description,
		"dependencies": pv.Dependencies,
		"sha256":       pv.SHA256,
		"size":         pv.Size,
		"published_at": pv.PublishedAt,
		"metadata":     pv.Metadata,
	}
}

// MapToPackage converts map to Package struct
func MapToPackage(m map[string]interface{}) *Package {
	p := &Package{}
	
	if name, ok := m["name"].(string); ok {
		p.Name = name
	}
	if desc, ok := m["description"].(string); ok {
		p.Description = desc
	}
	if latest, ok := m["latest"].(string); ok {
		p.Latest = latest
	}
	if versions, ok := m["versions"].([]interface{}); ok {
		p.Versions = make([]string, len(versions))
		for i, v := range versions {
			if str, ok := v.(string); ok {
				p.Versions[i] = str
			}
		}
	}
	if tags, ok := m["tags"].([]interface{}); ok {
		p.Tags = make([]string, len(tags))
		for i, v := range tags {
			if str, ok := v.(string); ok {
				p.Tags[i] = str
			}
		}
	}
	if updatedAt, ok := m["updated_at"].(time.Time); ok {
		p.UpdatedAt = updatedAt
	}
	
	return p
}

// MapToPackageVersion converts map to PackageVersion struct
func MapToPackageVersion(m map[string]interface{}) *PackageVersion {
	pv := &PackageVersion{}
	
	if name, ok := m["name"].(string); ok {
		pv.Name = name
	}
	if version, ok := m["version"].(string); ok {
		pv.Version = version
	}
	if desc, ok := m["description"].(string); ok {
		pv.Description = desc
	}
	if deps, ok := m["dependencies"].(map[string]interface{}); ok {
		pv.Dependencies = make(map[string]string)
		for k, v := range deps {
			if str, ok := v.(string); ok {
				pv.Dependencies[k] = str
			}
		}
	}
	if sha256, ok := m["sha256"].(string); ok {
		pv.SHA256 = sha256
	}
	if size, ok := m["size"].(int64); ok {
		pv.Size = size
	}
	if publishedAt, ok := m["published_at"].(time.Time); ok {
		pv.PublishedAt = publishedAt
	}
	if metadata, ok := m["metadata"].(map[string]interface{}); ok {
		pv.Metadata = metadata
	}
	
	return pv
}