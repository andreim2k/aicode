# Changelog

## [Unreleased] - Major Refactoring

### Fixed - Critical Issues
- ✅ **Eliminated code duplication**: Created shared `proxy` package used by both x-ai-proxy and z-ai-proxy
- ✅ **Fixed token exposure**: Tokens now use environment variables instead of CLI flags (not visible in `ps aux`)
- ✅ **Added authentication**: Proxy endpoints now support optional authentication middleware
- ✅ **Fixed graceful shutdown**: Uses `http.Server.Shutdown()` instead of `os.Exit()` for proper cleanup
- ✅ **Added HTTP timeouts**: All HTTP clients now have proper timeout configuration

### Fixed - High Priority Issues
- ✅ **Added input validation**: Request size limits, model validation, message count limits
- ✅ **Fixed error handling**: All JSON encoding errors are now properly handled
- ✅ **Fixed race conditions**: Improved cleanup functions with proper process management
- ✅ **Fixed port conflicts**: Better port availability checking and cleanup
- ✅ **Fixed naming conventions**: Removed `Z_AI` naming, using proper Go conventions

### Improved - Medium Priority Issues
- ✅ **Added health check endpoint**: `/health` endpoint for monitoring
- ✅ **Improved string concatenation**: Uses `strings.Join` for efficiency
- ✅ **Added request ID/tracing**: All requests now have correlation IDs
- ✅ **Added structured logging**: Better logging with request IDs
- ✅ **Added go.mod**: Proper Go module management

### Changed
- Proxy binaries now use environment variables (`XAI_TOKEN`, `ZAI_TOKEN`) instead of CLI flags for tokens
- Both proxies share the same codebase via the `proxy` package
- Improved error messages and validation
- Better process cleanup and port management

### Removed
- Removed unused `z-ai-adapter.sh` file

### Security
- Tokens are no longer visible in process lists
- Optional proxy authentication available
- Input validation prevents DoS attacks
- Proper error message sanitization
