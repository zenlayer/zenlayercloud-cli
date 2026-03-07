package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	// arrayFlagNames tracks which flags are array types for args expansion.
	arrayFlagNames []string
}

// stringSliceValue implements pflag.Value for space-separated string lists.
// Supports: --flag a b c (trailing args), --flag a --flag b (repeated)
type stringSliceValue struct {
	value   *[]string
	changed bool
}

func newStringSliceValue(p *[]string) *stringSliceValue {
	return &stringSliceValue{value: p}
}

func (s *stringSliceValue) Set(val string) error {
	if !s.changed {
		*s.value = []string{}
		s.changed = true
	}
	*s.value = append(*s.value, val)
	return nil
}

func (s *stringSliceValue) Type() string { return "strings" }

func (s *stringSliceValue) String() string {
	if s.value == nil || len(*s.value) == 0 {
		return ""
	}
	return strings.Join(*s.value, " ")
}

// intSliceValue implements pflag.Value for space-separated int lists.
type intSliceValue struct {
	value   *[]int
	changed bool
}

func newIntSliceValue(p *[]int) *intSliceValue {
	return &intSliceValue{value: p}
}

func (s *intSliceValue) Set(val string) error {
	if !s.changed {
		*s.value = []int{}
		s.changed = true
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid integer %q", val)
	}
	*s.value = append(*s.value, n)
	return nil
}

func (s *intSliceValue) Type() string { return "ints" }

func (s *intSliceValue) String() string {
	if s.value == nil || len(*s.value) == 0 {
		return ""
	}
	strs := make([]string, len(*s.value))
	for i, v := range *s.value {
		strs[i] = strconv.Itoa(v)
	}
	return strings.Join(strs, " ")
}

func newFlagStore() *flagStore {
	return &flagStore{
		strings:        make(map[string]*string),
		stringArrays:   make(map[string]*[]string),
		ints:           make(map[string]*int),
		intArrays:      make(map[string]*[]int),
		floats:         make(map[string]*float64),
		bools:          make(map[string]*bool),
		arrayFlagNames: []string{},
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
			enumVals := param.EnumValues.Values()
			cmd.RegisterFlagCompletionFunc(name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return enumVals, cobra.ShellCompDirectiveNoFileComp
			})
		} else {
			cmd.RegisterFlagCompletionFunc(name, noFileComp)
		}

	case "string-array":
		v := new([]string)
		store.stringArrays[name] = v
		store.arrayFlagNames = append(store.arrayFlagNames, name)
		cmd.Flags().Var(newStringSliceValue(v), name, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "integer":
		v := new(int)
		store.ints[name] = v
		cmd.Flags().IntVar(v, name, 0, desc)
		cmd.RegisterFlagCompletionFunc(name, noFileComp)

	case "integer-array":
		v := new([]int)
		store.intArrays[name] = v
		store.arrayFlagNames = append(store.arrayFlagNames, name)
		cmd.Flags().Var(newIntSliceValue(v), name, desc)
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
		store.arrayFlagNames = append(store.arrayFlagNames, name)
		cmd.Flags().Var(newStringSliceValue(v), name, desc)
		keys := extractSchemaKeys(param.ItemSchema)
		cmd.RegisterFlagCompletionFunc(name, schemaKeyComp(keys))
	}

	if param.Deprecated {
		cmd.Flags().MarkDeprecated(name, fmt.Sprintf("use a replacement flag instead. %s", param.Description))
	}
}

func buildFlagDescription(param *Parameter) string {
	desc := param.Description
	if idx := strings.Index(desc, "\n"); idx != -1 {
		desc = desc[:idx]
	}

	switch param.Type {
	case "enum":
		if len(param.EnumValues) > 0 {
			desc = fmt.Sprintf("%s (%s)", desc, strings.Join(param.EnumValues.Values(), "|"))
		}
	case "string-array":
		desc += " (one or more strings separated by spaces, quote items containing spaces)"
	case "integer-array":
		desc += " (one or more integers separated by spaces)"
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

// expandTrailingArgs processes positional arguments and appends them to the
// appropriate array flag. It finds the last array flag in os.Args and appends
// trailing args to that flag's value.
func expandTrailingArgs(cmd *cobra.Command, args []string, store *flagStore) error {
	if len(args) == 0 {
		return nil
	}

	// Find the last array flag that was specified in the command line.
	// We need to determine which flag these trailing args belong to.
	lastArrayFlag := findLastArrayFlag(cmd, store.arrayFlagNames)
	if lastArrayFlag == "" {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	// Directly append to the store's slice (more reliable than going through pflag.Value).
	if strSlice, ok := store.stringArrays[lastArrayFlag]; ok {
		*strSlice = append(*strSlice, args...)
		return nil
	}

	if intSlice, ok := store.intArrays[lastArrayFlag]; ok {
		for _, arg := range args {
			n, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid integer %q for --%s", arg, lastArrayFlag)
			}
			*intSlice = append(*intSlice, n)
		}
		return nil
	}

	return fmt.Errorf("internal error: flag %s not found in store", lastArrayFlag)
}

// findLastArrayFlag finds the last array-type flag that was set on the command line.
func findLastArrayFlag(cmd *cobra.Command, arrayFlagNames []string) string {
	if len(arrayFlagNames) == 0 {
		return ""
	}

	// Build a set of array flag names for quick lookup.
	arraySet := make(map[string]bool, len(arrayFlagNames))
	for _, name := range arrayFlagNames {
		arraySet[name] = true
	}

	// Check which array flags were actually changed (set by user).
	var lastChanged string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed && arraySet[f.Name] {
			lastChanged = f.Name
		}
	})

	return lastChanged
}

func schemaKeyComp(keys []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return keys, cobra.ShellCompDirectiveNoSpace
	}
}
