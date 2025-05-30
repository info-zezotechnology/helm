/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"helm.sh/helm/v4/pkg/plugin"
)

type pluginUninstallOptions struct {
	names []string
}

func newPluginUninstallCmd(out io.Writer) *cobra.Command {
	o := &pluginUninstallOptions{}

	cmd := &cobra.Command{
		Use:     "uninstall <plugin>...",
		Aliases: []string{"rm", "remove"},
		Short:   "uninstall one or more Helm plugins",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return compListPlugins(toComplete, args), cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			return o.complete(args)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run(out)
		},
	}
	return cmd
}

func (o *pluginUninstallOptions) complete(args []string) error {
	if len(args) == 0 {
		return errors.New("please provide plugin name to uninstall")
	}
	o.names = args
	return nil
}

func (o *pluginUninstallOptions) run(out io.Writer) error {
	slog.Debug("loading installer plugins", "dir", settings.PluginsDirectory)
	plugins, err := plugin.FindPlugins(settings.PluginsDirectory)
	if err != nil {
		return err
	}
	var errorPlugins []error
	for _, name := range o.names {
		if found := findPlugin(plugins, name); found != nil {
			if err := uninstallPlugin(found); err != nil {
				errorPlugins = append(errorPlugins, fmt.Errorf("failed to uninstall plugin %s, got error (%v)", name, err))
			} else {
				fmt.Fprintf(out, "Uninstalled plugin: %s\n", name)
			}
		} else {
			errorPlugins = append(errorPlugins, fmt.Errorf("plugin: %s not found", name))
		}
	}
	if len(errorPlugins) > 0 {
		return errors.Join(errorPlugins...)
	}
	return nil
}

func uninstallPlugin(p *plugin.Plugin) error {
	if err := os.RemoveAll(p.Dir); err != nil {
		return err
	}
	return runHook(p, plugin.Delete)
}

func findPlugin(plugins []*plugin.Plugin, name string) *plugin.Plugin {
	for _, p := range plugins {
		if p.Metadata.Name == name {
			return p
		}
	}
	return nil
}
