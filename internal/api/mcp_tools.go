package api

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

// registerMCPReadTools adds the tools every token may call.
//
// The tools themselves land in a follow-up; this file fixes the split that the
// read-only flag turns on, so the transport can be reviewed on its own.
func registerMCPReadTools(_ *mcp.Server, _ *gorm.DB) {}

// registerMCPWriteTools adds the tools withheld from read-only tokens.
func registerMCPWriteTools(_ *mcp.Server, _ *gorm.DB) {}
