// Package cli contains the Cobra adapter layer for chilly.
//
// The package stays in one directory on purpose:
//   - command entrypoints live in command-shaped files such as auth.go or user.go
//   - shared command support uses role-shaped files such as output_pretty.go
//     and schema_registry.go
//   - reusable transport, config, build, and update logic live outside this
//     package under internal/
//
// That keeps the public CLI surface easy to scan without forcing shallow
// subpackages that only add import churn.
package cli
