//go:build ruleguard

// Package gorules holds custom go-ruleguard rules executed via gocritic's
// `ruleguard` checker. They encode go-dev conventions that no off-the-shelf
// golangci-lint linter enforces.
//
// Wiring (see .golangci.yml):
//
//	linters.settings.gocritic.settings.ruleguard.rules: "ruleguard/rules.go"
//
// The build tag keeps this file out of normal `go build`; ruleguard parses it
// directly. Consumers copying .golangci.yml must also copy this file and keep
// the relative `rules:` path correct for their working directory.
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

// errVarNaming enforces: a variable whose type is error must be named `err`,
// never a descriptive name like taskErr / enqueueErr. Wrapping with
// fmt.Errorf supplies the context you would otherwise encode in the name.
func errVarNaming(m dsl.Matcher) {
	// single short var decl: `x := f()`
	m.Match(`$name := $_`).
		Where(m["name"].Type.Is(`error`) &&
			m["name"].Text != "err" &&
			!m["name"].Text.Matches(`^_$`)).
		At(m["name"]).
		Report(`error-typed variable should be named "err", not $name`)

	// two-value short var decl: `v, e := f()` (catches `task, taskErr := ...`)
	m.Match(`$_, $name := $_`, `$name, $_ := $_`).
		Where(m["name"].Type.Is(`error`) &&
			m["name"].Text != "err" &&
			!m["name"].Text.Matches(`^_$`)).
		At(m["name"]).
		Report(`error-typed variable should be named "err", not $name`)

	// explicit declaration: `var taskErr error`
	m.Match(`var $name error`).
		Where(m["name"].Text != "err" &&
			!m["name"].Text.Matches(`^_$`)).
		At(m["name"]).
		Report(`error-typed variable should be named "err", not $name`)
}

// errStringStyle enforces concise error strings: no "failed to" prefix.
// Good:  fmt.Errorf("create task: %w", err)
// Bad:   fmt.Errorf("failed to create task: %w", err)
func errStringStyle(m dsl.Matcher) {
	m.Match(`fmt.Errorf($msg, $*_)`, `errors.New($msg)`).
		Where(m["msg"].Text.Matches(`(?i)^.failed to`)).
		At(m["msg"]).
		Report(`avoid "failed to" in error strings; use concise context, e.g. "create task: %w"`)
}

func noMultiVlaueInlineErr(m dsl.Matcher) {
	m.Match(`if $val, $err := $_; $_ { $*_ }`,
		`if $val, $err := $_; $_ { $*_ } else { $*_ }`).
		Where(m["err"].Type.Is("error") && m["val"], Text != "_").
		Report(`avoid multi-value inline if-init with named value: assign on its own line, then check err`)
}
