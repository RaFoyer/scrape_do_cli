package cmd

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/ra/scrape_do_cli/internal/outfmt"
)

type SchemaCmd struct {
	Command       []string `arg:"" optional:"" name:"command" help:"Optional command path to describe"`
	IncludeHidden bool     `name:"include-hidden" help:"Include hidden commands and flags"`
}

type schemaDoc struct {
	SchemaVersion int         `json:"schema_version"`
	Build         string      `json:"build"`
	Command       *schemaNode `json:"command"`
}

type schemaNode struct {
	Type        string        `json:"type"`
	Name        string        `json:"name"`
	Aliases     []string      `json:"aliases,omitempty"`
	Help        string        `json:"help,omitempty"`
	Path        string        `json:"path"`
	Usage       string        `json:"usage,omitempty"`
	Hidden      bool          `json:"hidden,omitempty"`
	Flags       []schemaFlag  `json:"flags,omitempty"`
	Subcommands []*schemaNode `json:"subcommands,omitempty"`
}

type schemaFlag struct {
	Name       string   `json:"name"`
	Aliases    []string `json:"aliases,omitempty"`
	Short      string   `json:"short,omitempty"`
	Help       string   `json:"help,omitempty"`
	Type       string   `json:"type"`
	Required   bool     `json:"required,omitempty"`
	Default    string   `json:"default,omitempty"`
	HasDefault bool     `json:"has_default,omitempty"`
	Hidden     bool     `json:"hidden,omitempty"`
}

func (c *SchemaCmd) Run(ctx context.Context, kctx *kong.Context) error {
	ctx = outfmt.WithJSONTransform(ctx, outfmt.JSONTransform{})
	root := kctx.Model.Node
	node := root
	cmdPath := splitCommandPath(c.Command)
	if len(cmdPath) > 0 {
		found, err := findCommandNode(root, cmdPath)
		if err != nil {
			return err
		}
		node = found
	}
	hide := !c.IncludeHidden
	doc := schemaDoc{
		SchemaVersion: 1,
		Build:         VersionString(),
		Command:       buildSchemaNode(node, hide),
	}
	return outfmt.WriteJSON(ctx, os.Stdout, doc)
}

func splitCommandPath(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		for _, tok := range strings.Fields(strings.TrimSpace(p)) {
			if tok != "" {
				out = append(out, tok)
			}
		}
	}
	return out
}

func findCommandNode(root *kong.Node, path []string) (*kong.Node, error) {
	cur := root
	for _, token := range path {
		next := findChildCommand(cur, token)
		if next == nil {
			return nil, usagef("unknown command %q under %q", token, strings.TrimSpace(cur.FullPath()))
		}
		cur = next
	}
	return cur, nil
}

func findChildCommand(parent *kong.Node, token string) *kong.Node {
	token = strings.ToLower(strings.TrimSpace(token))
	for _, child := range parent.Children {
		if child == nil || child.Type != kong.CommandNode {
			continue
		}
		if strings.ToLower(child.Name) == token {
			return child
		}
		for _, alias := range child.Aliases {
			if strings.ToLower(strings.TrimSpace(alias)) == token {
				return child
			}
		}
	}
	return nil
}

func buildSchemaNode(node *kong.Node, hide bool) *schemaNode {
	if node == nil {
		return nil
	}
	out := &schemaNode{
		Type:    schemaNodeType(node),
		Name:    node.Name,
		Aliases: sortedStrings(node.Aliases),
		Help:    strings.TrimSpace(node.Help),
		Path:    strings.TrimSpace(node.FullPath()),
		Usage:   strings.TrimSpace(node.Summary()),
		Hidden:  node.Hidden,
		Flags:   schemaFlags(node, hide),
	}
	children := make([]*kong.Node, 0, len(node.Children))
	for _, child := range node.Children {
		if child == nil || child.Type != kong.CommandNode {
			continue
		}
		if hide && child.Hidden {
			continue
		}
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool { return children[i].Name < children[j].Name })
	for _, child := range children {
		out.Subcommands = append(out.Subcommands, buildSchemaNode(child, hide))
	}
	return out
}

func schemaNodeType(node *kong.Node) string {
	switch node.Type {
	case kong.ApplicationNode:
		return "application"
	case kong.CommandNode:
		return "command"
	case kong.ArgumentNode:
		return "argument"
	default:
		return "unknown"
	}
}

func schemaFlags(node *kong.Node, hide bool) []schemaFlag {
	out := []schemaFlag{}
	for _, group := range node.AllFlags(hide) {
		for _, f := range group {
			if f == nil {
				continue
			}
			out = append(out, schemaFlag{
				Name:       f.Name,
				Aliases:    sortedStrings(f.Aliases),
				Short:      flagShortString(f.Short),
				Help:       strings.TrimSpace(f.Help),
				Type:       reflectTypeString(f.Target),
				Required:   f.Required,
				Default:    strings.TrimSpace(f.Default),
				HasDefault: f.HasDefault,
				Hidden:     f.Hidden,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func reflectTypeString(v any) string {
	if v == nil {
		return "unknown"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		return "[]" + t.Elem().String()
	}
	return t.String()
}

func flagShortString(r rune) string {
	if r == 0 {
		return ""
	}
	return string(r)
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
