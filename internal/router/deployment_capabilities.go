package router

import (
	"github.com/ai-tool-collection/WeKnora/internal/handler"
	"github.com/ai-tool-collection/WeKnora/internal/sandbox"
)

func deploymentCapabilitiesFromRouter(params RouterParams) handler.DeploymentCapabilitiesData {
	return handler.BuildDeploymentCapabilities(handler.Edition, handler.DeploymentFeatureAvailability{
		Organizations: params.OrganizationHandler != nil,
		Agents:        params.CustomAgentHandler != nil,
		IM:            params.IMHandler != nil,
		// Match RegisterEmbedChannelRoutes: management routes depend on handler only.
		Embed: params.EmbedChannelHandler != nil,
		// Match RegisterMCPEndpointRoutes / RegisterMCPServerRoutes.
		MCPServer: params.MCPEndpointHandler != nil && params.MCPServer != nil && params.MCPEndpointService != nil,
		API:       params.TenantHandler != nil && params.TenantAPIKeyService != nil,
		MCP: params.MCPServiceHandler != nil &&
			params.MCPCredentialsHandler != nil &&
			params.MCPOAuthHandler != nil,
		WebSearch: params.WebSearchHandler != nil &&
			params.WebSearchProviderHandler != nil &&
			params.WebSearchCredentialsHandler != nil,
		VectorStore:   params.VectorStoreHandler != nil,
		Storage:       params.StorageBackendHandler != nil,
		Sandbox:       params.SandboxConfigHandler != nil,
		SandboxDocker: sandbox.DockerBackendEnabled(),
		SandboxHost:   params.HostSandbox.Manager != nil,
		SandboxRemote: params.SandboxConfigHandler != nil && !params.HostSandbox.Desktop,
	})
}
