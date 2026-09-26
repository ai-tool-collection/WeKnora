#!/usr/bin/env python3
"""Start the Knowledge Hub MCP server. Diagnostics go to stderr for stdio MCP."""

import argparse
import asyncio
import logging
import os
import sys
from pathlib import Path


def setup_environment():
    current_dir = str(Path(__file__).parent.absolute())
    if current_dir not in sys.path:
        sys.path.insert(0, current_dir)


def check_dependencies():
    try:
        import mcp  # noqa: F401
        import requests  # noqa: F401
    except ImportError as exc:
        print(f"Missing dependency: {exc}", file=sys.stderr)
        print("Install dependencies: pip install -r requirements.txt", file=sys.stderr)
        return False
    return True


def check_environment_variables():
    base_url = os.getenv("WEKNORA_BASE_URL")
    api_key = os.getenv("WEKNORA_API_KEY")
    print("=== Knowledge Hub MCP Server environment ===", file=sys.stderr)
    print(f"Base URL: {base_url or 'http://localhost:8080/api/v1 (default)'}", file=sys.stderr)
    print(f"API Key: {'configured' if api_key else 'not configured'}", file=sys.stderr)
    if not base_url:
        print("Set WEKNORA_BASE_URL to use a different server.", file=sys.stderr)
    if not api_key:
        print("Set WEKNORA_API_KEY to authenticate requests.", file=sys.stderr)
    print("=" * 40, file=sys.stderr)
    return True


def parse_arguments():
    parser = argparse.ArgumentParser(
        description="Knowledge Hub MCP Server",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""Examples:
  python main.py
  python main.py --check-only
  python main.py --transport http --host 127.0.0.1 --port 8000

Environment: WEKNORA_BASE_URL, WEKNORA_API_KEY, MCP_SERVER_AUTH_TOKEN
""",
    )
    parser.add_argument("--check-only", action="store_true", help="Check the environment and exit")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable debug logging")
    parser.add_argument("--version", action="version", version="Knowledge Hub MCP Server 1.1.1")
    parser.add_argument("--transport", choices=["stdio", "sse", "http"],
                        default=os.getenv("MCP_TRANSPORT", "stdio"))
    parser.add_argument("--host", default=os.getenv("MCP_HOST", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("MCP_PORT", "8000")))
    return parser.parse_args()


async def main():
    args = parse_arguments()
    setup_environment()
    if not check_dependencies():
        raise SystemExit(1)
    check_environment_variables()
    if args.check_only:
        print("Environment check complete.", file=sys.stderr)
        return
    if args.verbose:
        logging.basicConfig(level=logging.DEBUG)
    try:
        print(f"Starting Knowledge Hub MCP Server ({args.transport})...", file=sys.stderr)
        from weknora_mcp_server import run_stdio, run_sse, run_http
        if args.transport == "stdio":
            await run_stdio()
        elif args.transport == "sse":
            await run_sse(args.host, args.port)
        else:
            await run_http(args.host, args.port)
    except ImportError as exc:
        print(f"Import error: {exc}", file=sys.stderr)
        raise SystemExit(1) from exc
    except KeyboardInterrupt:
        print("\nServer stopped", file=sys.stderr)
    except Exception as exc:
        print(f"Server error: {exc}", file=sys.stderr)
        if args.verbose:
            logging.exception("MCP server failed")
        raise SystemExit(1) from exc


def sync_main():
    asyncio.run(main())


if __name__ == "__main__":
    sync_main()
