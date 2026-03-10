package loader

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// flagStore holds runtime values for all dynamically registered flags of one command.
// Keyed by parameter name (kebab-case).
type flagStore struct {
	strings      map[string]*string
	stringArrays map[string]*[]string
	ints         map[string]*int
	intArrays    map[string]*[]int
	floats       map[string]*float64
	bools        map[string]*bool
}

func newFlagStore() *flagStore {
	return &flagStore{
		strings:      make(map[string]*string),
		stringArrays: make(map[string]*[]string),
		ints:         make(map[string]*int),
		intArrays:    make(map[string]*[]int),
		floats:       make(map[string]*float64),
		bools:        make(map[string]*bool),
	}
}

// bindFlags registers all parameters of a definition as cobra flags, returning
// the flag store that holds the bound variables.
func bindFlags(cmd *cobra.Command, def *APIDefinition) *flagStore {
	store := newFlagStore()
	for i := range def.Parameters {
		param := &def.Parameters[i]
		bindFlag(cmd, param, store)
	}
	return store
}

func bindFlag(cmd *cobra.Command, param *Parameter, store *flagStore) {
	name := param.Name
	desc := buildFlagDescription(param)

	switch param.Type {
	case "string", "enum":
		v := new(string)
		store.strings[name] = v
		cmd.Flags().StringVar(v, name, "", desc)

		if param.Type == "enum" {
			cmd.RegisterFlagCompletionFunc(name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return param.EnumValues, cobra.ShellCompDirectiveNoFileComp
			})
		} else {
			cmd.RegisterFlagCompletionFunc(name, noFileComp)
		}

	case "string-array":
		v := new([]string)
		store.stringArrays[name] = v
		cmd.Flags().StringArrayVar(v, name, nil, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "integer":
		v := new(int)
		store.ints[name] = v
		cmd.Flags().IntVar(v, name, 0, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "integer-array":
		v := new([]int)
		store.intArrays[name] = v
		cmd.Flags().IntSliceVar(v, name, nil, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "float":
		v := new(float64)
		store.floats[name] = v
		cmd.Flags().Float64Var(v, name, 0, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "boolean":
		v := new(bool)
		store.bools[name] = v
		cmd.Flags().BoolVar(v, name, false, desc)
		// cobra automatically completes true/false for BoolVar

	case "object":
		// Stored as a string (kv or JSON), parsed at runtime.
		v := new(string)
		store.strings[name] = v
		cmd.Flags().StringVar(v, name, "", desc)
		keys := extractSchemaKeys(param.ObjectSchema)
		cmd.RegisterFlagCompletionFunc(name, schemaKeyComp(keys))

	case "object-array":
		v := new([]string)
		store.stringArrays[name] = v
		cmd.Flags().StringArrayVar(v, name, nil, desc)
		keys := extractSchemaKeys(param.ItemSchema)
		cmd.RegisterFlagCompletionFunc(name, schemaKeyComp(keys))
	}

	if param.Deprecated {
		cmd.Flags().MarkDeprecated(name, fmt.Sprintf("use a replacement flag instead. %s", param.Description))
	}
}

func buildFlagDescription(param *Parameter) string {
	desc := param.Description

	switch param.Type {
	case "enum":
		if len(param.EnumValues) > 0 {
			desc = fmt.Sprintf("%s (%s)", desc, strings.Join(param.EnumValues, "|"))
		}
	case "object":
		if param.ArrayStyle == "kv" || param.ArrayStyle == "" {
			parts := objectExampleParts(param.ObjectSchema)
			if len(parts) > 0 {
				desc = fmt.Sprintf("%s (e.g. %s)", desc, strings.Join(parts, ","))
			}
		}
	case "object-array":
		if param.ArrayStyle == "kv" || param.ArrayStyle == "" {
			parts := objectExampleParts(param.ItemSchema)
			if len(parts) > 0 {
				desc = fmt.Sprintf("%s (e.g. %s)", desc, strings.Join(parts, ","))
			}
		}
	}

	if param.Sensitive {
		desc += " (sensitive)"
	}

	return desc
}

func extractSchemaKeys(fields []SchemaField) []string {
	keys := make([]string, 0, len(fields))
	for _, f := range fields {
		keys = append(keys, f.Name+"=")
	}
	return keys
}

// objectExampleParts builds example key=<type> strings for description hints.
func objectExampleParts(fields []SchemaField) []string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		placeholder := "<value>"
		switch f.Type {
		case "integer":
			placeholder = "<int>"
		case "float":
			placeholder = "<float>"
		case "boolean":
			placeholder = "<bool>"
		}
		parts = append(parts, f.Name+"="+placeholder)
	}
	return parts
}

func noFileComp(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func schemaKeyComp(keys []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return keys, cobra.ShellCompDirectiveNoSpace
	}
}
