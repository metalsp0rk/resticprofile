package config

import (
	"bytes"
	"time"
)

// Helpers for tests

func getProfile(configFormat, configString, profileKey string) (*Profile, error) {
	c, err := Load(bytes.NewBufferString(configString), configFormat)
	if err != nil {
		return nil, err
	}

	profile, err := c.GetProfile(profileKey)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func getResolvedProfile(configFormat, configString, profileKey string) (*Profile, error) {
	profile, err := getProfile(configFormat, configString, profileKey)
	if err != nil {
		return nil, err
	}

	data := TemplateData{
		Profile: ProfileTemplateData{
			Name: profile.Name,
		},
		Now:        time.Now(),
		ConfigDir:  "ConfigDir",
		CurrentDir: "CurrentDir",
	}
	err = profile.ResolveTemplates(data)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
