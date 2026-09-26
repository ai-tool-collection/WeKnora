#!/usr/bin/env python3
"""
Knowledge Hub MCP Server Package

A Model Context Protocol server that provides access to the Knowledge Hub API.
"""

__version__ = "1.1.1"
__author__ = "Knowledge Hub Contributors"
__description__ = "Knowledge Hub MCP Server - Model Context Protocol server for Knowledge Hub API"

from weknora_mcp_server import WeKnoraClient, run

__all__ = ["WeKnoraClient", "run"]
