/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package app

import (
	"fmt"
	"github.com/compose-spec/compose-go/types"
	"gopkg.in/yaml.v2"
	"strings"
)

type ComposeBuilder struct {
	Manifest Manifest
}

// newComposeBuilder called internally by NewBuilder()
func newComposeBuilder(appMan Manifest) Builder {
	return &ComposeBuilder{Manifest: appMan}
}

func (b *ComposeBuilder) Build() ([]DeploymentRsx, error) {
	p := b.buildProject()
	composeProject, err := yaml.Marshal(p)
	if err != nil {
		return nil, err
	}
	return []DeploymentRsx{
		{
			Name:    "docker-compose.yml",
			Content: composeProject,
			Type:    ComposeProject,
		},
	}, nil
}

func (b *ComposeBuilder) buildProject() types.Project {
	p := types.Project{}
	p.Name = fmt.Sprintf("Docker Compose Project for %s", strings.ToUpper(b.Manifest.Name))
	for _, svc := range b.Manifest.Services {
		p.Services = append(p.Services, types.ServiceConfig{
			Name:          svc.Name,
			ContainerName: svc.Name,
			DependsOn:     getDeps(svc.DependsOn),
			Environment:   getEnv(svc.Info.Var),
			Image:         svc.Image,
			Ports:         nil,
			Restart:       "always",
			Volumes:       getSvcVols(svc.Info.Volume),
		})
	}
	p.Volumes = getVols(b.Manifest.Services)
	return p
}

func getSvcVols(volume []Volume) []types.ServiceVolumeConfig {
	vo := make([]types.ServiceVolumeConfig, 0)
	for _, v := range volume {
		vo = append(vo, types.ServiceVolumeConfig{
			Extensions: map[string]interface{}{
				v.Name: v.Path,
			},
		})
	}
	return vo
}

func getDeps(dependencies []string) types.DependsOnConfig {
	d := types.DependsOnConfig{}
	for _, dependency := range dependencies {
		d[dependency] = types.ServiceDependency{Condition: types.ServiceConditionStarted}
	}
	return d
}

func getEnv(vars []Var) types.MappingWithEquals {
	var values []string
	for _, v := range vars {
		values = append(values, fmt.Sprintf("%s=%s", v.Name, v.Value))
	}
	return types.NewMappingWithEquals(values)
}

func getVols(svc []SvcRef) types.Volumes {
	vo := types.Volumes{}
	for _, s := range svc {
		for _, v := range s.Info.Volume {
			vo[v.Name] = types.VolumeConfig{
				External: types.External{External: true},
			}
		}
	}
	return vo
}
