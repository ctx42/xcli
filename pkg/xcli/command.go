// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/ctx42/ring/pkg/ring"
	"github.com/ctx42/xflag/pkg/xflag"
)

// PreRun is the signature of functions run before a command. It takes the
// context and the command context, and returns a potentially modified context
// and an error. The caller must use the returned context.
type PreRun func(ctx context.Context, ccx *CmdCtx) (context.Context, error)

// WithFlags is option for [NewCommand] setting flag set to use instead of
// default empty flag set.
//
// Unlike the default flag set, the given one keeps its Usage function and
// output, so on a parse error or a help request [flag.FlagSet] writes to its
// output (standard error when not set) instead of the [ring.Ring] streams.
// Set Usage to nil to have [Command.Run] print [Command.Help] instead.
func WithFlags(fs *xflag.FlagSet) func(*Command) {
	return func(cmd *Command) { cmd.flags = fs }
}

// Command represents CLI command and its sub-commands.
//
// The zero value is not usable. Use [NewCommand] to create a command.
type Command struct {
	// Metadata shared by all commands in the tree.
	meta Meta

	// Command flags.
	flags *xflag.FlagSet

	// Parent command (nil for root command).
	parent *Command

	// List of subcommands.
	subs []*Command

	// Function run before the fn is called.
	pre []PreRun

	// The function that encapsulates the command logic.
	fn Func

	// Arguments validation.
	argsRule Rule

	// After parsing holds the non-subcommand and non-flag arguments.
	args []string

	// Usage flag name.
	usageFlag string

	// Set when parsing met -h or --help with no such flag defined.
	helpReq bool

	// Short, one line, description.
	shortDesc string

	// Long, multi line, description.
	longDesc string

	// Usage examples and their descriptions.
	examples [][2]string

	// Set on the root command by Parse to the command Run runs.
	found *Command
}

// NewCommand instantiates new command.
func NewCommand(name string, fn Func, opts ...func(*Command)) *Command {
	cmd := &Command{
		meta:     NewMeta(),
		argsRule: ArgsRule(0),
		fn:       fn,
	}
	for _, opt := range opts {
		opt(cmd)
	}
	if cmd.flags == nil {
		cmd.flags = xflag.NewFlagSet(name, flag.ContinueOnError)
		cmd.flags.Usage = nil
		cmd.flags.SetOutput(io.Discard)
	} else {
		// The flag set must have command's name.
		cmd.flags.Init(name, cmd.flags.ErrorHandling())
	}
	return cmd
}

// Name returns the command name.
func (cmd *Command) Name() string { return cmd.flags.Name() }

// Flags returns the command flag set.
func (cmd *Command) Flags() *xflag.FlagSet { return cmd.flags }

// MetaLookup returns the value of the metadata variable named by the key and
// true, or nil and false when the variable is not set. The metadata is shared
// by all commands in the tree.
func (cmd *Command) MetaLookup(key string) (any, bool) {
	return cmd.meta.MetaLookup(key)
}

// MetaGet returns the value of the metadata variable named by the key, or nil
// when the variable is not set. The metadata is shared by all commands in the
// tree.
func (cmd *Command) MetaGet(key string) any { return cmd.meta.MetaGet(key) }

// MetaSet sets the value of the metadata variable named by the key. The
// metadata is shared by all commands in the tree.
func (cmd *Command) MetaSet(key string, value any) *Command {
	cmd.meta.MetaSet(key, value)
	return cmd
}

// MetaDelete deletes the metadata variable named by the key. The metadata is
// shared by all commands in the tree.
func (cmd *Command) MetaDelete(key string) { cmd.meta.MetaDelete(key) }

// MetaAll returns the map backing the metadata shared by all commands in the
// tree.
func (cmd *Command) MetaAll() map[string]any { return cmd.meta.MetaAll() }

// Add adds subcommand(s) to the current command.
func (cmd *Command) Add(cs ...*Command) *Command {
	for _, c := range cs {
		c.parent = cmd
		c.setMeta(cmd.meta)
		cmd.subs = append(cmd.subs, c)
	}
	return cmd
}

// setMeta sets the metadata collection on the command and all its
// subcommands at every depth.
func (cmd *Command) setMeta(meta Meta) {
	cmd.meta = meta
	for _, sub := range cmd.subs {
		sub.setMeta(meta)
	}
}

// UsageFlagName sets the name of the help (usage) flag.
func (cmd *Command) UsageFlagName(name string) *Command {
	cmd.usageFlag = name
	return cmd
}

// ShouldRunUsage returns true when Usage method should be run.
func (cmd *Command) ShouldRunUsage() bool {
	if cmd.helpReq {
		return true
	}
	if cmd.parent != nil && cmd.parent.ShouldRunUsage() {
		return true
	}
	return cmd.flags.GetBool(cmd.usageFlag)
}

// ShortDesc sets short description for the command.
func (cmd *Command) ShortDesc(desc string) *Command {
	cmd.shortDesc = desc
	return cmd
}

// LongDesc sets long description for the command.
func (cmd *Command) LongDesc(desc string) *Command {
	cmd.longDesc = desc
	return cmd
}

// Example adds usage example for the command. The `desc` will be shown as an
// example description when help is displayed.
func (cmd *Command) Example(eg, desc string) *Command {
	cmd.examples = append(cmd.examples, [...]string{eg, desc})
	return cmd
}

// ArgsRule sets argument count rules. A nil rule sets the default rule
// requiring no arguments.
func (cmd *Command) ArgsRule(rule Rule) *Command {
	if rule == nil {
		rule = ArgsRule(0)
	}
	cmd.argsRule = rule
	return cmd
}

// Args returns the non-subcommand and non-flag arguments.
func (cmd *Command) Args() []string {
	return cmd.args
}

// Parse parses arguments for the command and all subcommands.
//
// Example:
//
//	-a 1 sub-1 -s1a 2
//	sub-1 -s1a 2
func (cmd *Command) Parse(args ...string) error {
	cmd.args = args
	found, err := cmd.parse(args)
	if err != nil {
		return err
	}
	cmd.found = found
	cmd.args = found.args
	return found.checkRequired()
}

// parse parses arguments recursively for all the commands in the tree.
func (cmd *Command) parse(args []string) (*Command, error) {
	// The `parse` is called to find arguments for the current command up to
	// the first sub-command name or this command arguments.
	err := cmd.flags.Parse(args)

	// The [flag.ErrHelp] is returned when there is no formal -h or --help
	// flag defined. We record it, so the Run method displays help.
	cmd.helpReq = errors.Is(err, flag.ErrHelp)
	if err != nil && !cmd.helpReq {
		return nil, err
	}

	cmd.args = cmd.flags.Args()
	if len(cmd.args) == 0 {
		// The `parse` is called to find and parse the flags
		// for the next command, if there are no more arguments
		// there is no need to continue.
		return cmd, nil
	}

	for _, sub := range cmd.subs {
		if cmd.args[0] != sub.Name() {
			continue
		}
		// Found subcommand - call `parse` recursively and return.
		// There is no need to test for the other sub-commands of the current
		// command because the syntax does not allow to put two commands of
		// the same level one after the other. It's always like this:
		//
		//   ROOT -> sub1 -> sub2 -> sub3 -> ...
		//
		return sub.parse(cmd.args[1:])
	}

	// The remaining arguments are for the current command.
	return cmd, nil
}

// Find finds command by its path. The path may start with the name of the
// command itself; the remaining names are matched against subcommands.
func (cmd *Command) Find(path ...string) *Command {
	if len(path) > 0 && path[0] == cmd.Name() {
		path = path[1:]
	}
	return cmd.find(path)
}

// find finds the subcommand by its path relative to the command.
func (cmd *Command) find(path []string) *Command {
	if len(path) == 0 {
		return cmd
	}
	for _, sub := range cmd.subs {
		if sub.Name() == path[0] {
			return sub.find(path[1:])
		}
	}
	return nil
}

// PreRun sets the function which will be run before the command.
func (cmd *Command) PreRun(fns ...PreRun) *Command {
	cmd.pre = append(cmd.pre, fns...)
	return cmd
}

// Run runs either found or this command. If pre-run functions were defined,
// they will be run in the order from root command to the found one.
func (cmd *Command) Run(ctx context.Context, rng *ring.Ring) (err error) {
	if cmd.found == nil || !cmd.flags.Parsed() {
		return errors.New("not parsed arguments")
	}
	found := cmd.found
	ccx := NewCmdCtx(rng, found)
	if found.ShouldRunUsage() {
		if found.flags.Usage != nil {
			found.flags.Usage()
			return nil
		}
		_, _ = fmt.Fprint(rng.Stderr(), found.Help())
		return nil
	}
	if found.fn == nil && len(found.subs) > 0 && len(found.args) > 0 {
		return fmt.Errorf("%w: %#q", ErrUnkCommand, found.args[0])
	}
	if err = found.argsRule.Validate(ccx.Cmd.Args()); err != nil {
		return fmt.Errorf("command %#q %w", found.Name(), err)
	}
	if ctx, err = RunPreRun(ctx, ccx); err != nil {
		return err
	}

	fn := found.fn
	if fn == nil {
		if len(found.subs) > 0 {
			return fmt.Errorf("%w for %#q", ErrSubCmdReq, found.Name())
		}
		fn = defaultFunc
	}
	return fn(ctx, ccx)
}

// RunPreRun runs all the pre-run functions in the command chain starting at
// root.
func RunPreRun(ctx context.Context, ccx *CmdCtx) (context.Context, error) {
	var err error
	if ccx.Cmd.parent != nil {
		orig := ccx.Cmd
		ccx.Cmd = ccx.Cmd.parent
		ctx, err = RunPreRun(ctx, ccx)
		ccx.Cmd = orig
		if err != nil {
			return nil, err
		}
	}
	for _, pre := range ccx.Cmd.pre {
		if ctx, err = pre(ctx, ccx); err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

// GetFlagSet returns the flag set of the command or of its closest ancestor
// with the given name. Subcommands are not searched. If no such command
// exists, it returns an "empty" flag set, so you can still do chain method
// calls.
func (cmd *Command) GetFlagSet(name string) *xflag.FlagSet {
	if cmd.Name() == name {
		return cmd.flags
	}
	if cmd.parent == nil {
		fs := xflag.NewFlagSet(name, cmd.flags.ErrorHandling())
		fs.Usage = nil
		fs.SetOutput(cmd.flags.Output())
		return fs
	}
	return cmd.parent.GetFlagSet(name)
}

// IsRoot returns true if the command is a root command.
//
// Root command is a command which has no parent command.
func (cmd *Command) IsRoot() bool { return cmd.parent == nil }

// checkRequired checks all required flags were set. Returns an error when a
// required flag is not set. If usage (help) flag has been requested it ignores
// the required flags and returns nil.
func (cmd *Command) checkRequired() error {
	if err := cmd.flags.CheckRequired(); err != nil && !cmd.ShouldRunUsage() {
		return fmt.Errorf("%w for command %#q", err, cmd.Name())
	}
	if cmd.parent != nil {
		return cmd.parent.checkRequired()
	}
	return nil
}

// describe returns a string describing current state of all commands, their
// flags and values.
func (cmd *Command) describe() string {
	buf := &strings.Builder{}
	cmd.desc(0, buf)
	return buf.String()
}

// desc is called recursively on all commands and their flags to build a string
// representing the current status of the command(s).
func (cmd *Command) desc(padding int, buf *strings.Builder) {
	pad := strings.Repeat(" ", padding*2)
	var found string
	if cmd.found != nil {
		found = fmt.Sprintf("(%s)", cmd.found.Name())
	}
	_, _ = fmt.Fprintf(buf, "%scmd%s: %s", pad, found, cmd.Name())
	cmd.flags.VisitAll(func(flg *flag.Flag) {
		_, _ = fmt.Fprintf(buf, " --%s %v", flg.Name, flg.Value)
	})
	_, _ = buf.WriteString("\n")
	for _, sub := range cmd.subs {
		sub.desc(padding+1, buf)
	}
}

// FullName returns command name also considering the parent names.
func (cmd *Command) FullName() string {
	name := cmd.Name()
	if cmd.parent != nil {
		name = cmd.parent.FullName() + " " + name
	}
	return name
}

// Help returns the command's help string, with the padding tabwriter adds
// at the end of lines removed.
func (cmd *Command) Help() string {
	var buf strings.Builder
	header := fmt.Sprintf("Usage of %#q:\n", cmd.FullName())
	_, _ = buf.WriteString(header)
	if cmd.longDesc != "" {
		desc := fmt.Sprintf("\n  %s\n", cmd.longDesc)
		_, _ = buf.WriteString(desc)
	}
	if cmd.longDesc == "" && cmd.shortDesc != "" {
		desc := fmt.Sprintf("\n  %s\n", cmd.shortDesc)
		_, _ = buf.WriteString(desc)
	}
	if len(cmd.examples) > 0 {
		_, _ = buf.WriteString("\nEXAMPLES:\n")
		tw := tabwriter.NewWriter(&buf, 0, 8, 4, ' ', 0)
		for _, egd := range cmd.examples {
			eg := egd[0]
			if egd[1] != "" {
				eg += "\t# " + egd[1]
			} else {
				eg += "\t"
			}
			_, _ = fmt.Fprintf(tw, "  %s\n", eg)
		}
		_ = tw.Flush()
	}
	_, _ = buf.WriteString(cmd.help())
	lines := strings.Split(buf.String(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

// help returns command's options and subcommands summary.
func (cmd *Command) help() string {
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 0, 8, 4, ' ', 0)
	opts := cmd.flags.HelpOptionLines()
	if len(opts) > 0 {
		_, _ = fmt.Fprintf(tw, "\t\n")
		_, _ = fmt.Fprintf(tw, "OPTIONS:\t\n")
		for _, opt := range opts {
			_, _ = fmt.Fprint(tw, opt)
		}
	}

	if len(cmd.subs) > 0 {
		_, _ = fmt.Fprintf(tw, "\t\n")
		_, _ = fmt.Fprintf(tw, "COMMANDS:\t\n")
	}
	for _, sub := range cmd.subs {
		_, _ = fmt.Fprintf(tw, "  %s\t%s\n", sub.Name(), sub.shortDesc)
	}
	_ = tw.Flush()
	return buf.String()
}
