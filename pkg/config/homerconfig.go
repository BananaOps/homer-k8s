package config

import (
	homerv1alpha1 "github.com/BananaOps/homer-k8s/api/v1alpha1"
)

type HomerConfig struct {
	Title    string                `yaml:"title"`
	Subtitle string                `yaml:"subtitle"`
	Logo     string                `yaml:"logo"`
	Icon     string                `yaml:"icon"`
	Footer   string                `yaml:"footer"`
	Columns  string                `yaml:"columns"`
	Colors   Colors                `yaml:"colors"`
	Message  Message               `yaml:"message"`
	Links    []Link                `yaml:"links"`
	Services []homerv1alpha1.Group `yaml:"services"`
}

type Colors struct {
	Light Color `yaml:"light"`
	Dark  Color `yaml:"dark"`
}

type Color struct {
	HighlightPrimary   string `yaml:"highlight-primary"`
	HighlightSecondary string `yaml:"highlight-secondary"`
	HighlightHover     string `yaml:"highlight-hover"`
	Background         string `yaml:"background"`
	CardBackground     string `yaml:"card-background"`
	Text               string `yaml:"text"`
	TextHeader         string `yaml:"text-header"`
	TextTitle          string `yaml:"text-title"`
	TextSubtitle       string `yaml:"text-subtitle"`
	CardShadow         string `yaml:"card-shadow"`
	Link               string `yaml:"link"`
	LinkHover          string `yaml:"link-hover"`
	BackgroundImage    string `yaml:"background-image"`
}

type Message struct {
	URL             string            `yaml:"url"`
	Mapping         map[string]string `yaml:"mapping"`
	RefreshInterval int               `yaml:"refreshInterval"`
	Style           string            `yaml:"style"`
	Title           string            `yaml:"title"`
	Icon            string            `yaml:"icon"`
	Content         string            `yaml:"content"`
}

type Link struct {
	Name   string `yaml:"name"`
	Icon   string `yaml:"icon"`
	URL    string `yaml:"url"`
	Target string `yaml:"target"`
}
