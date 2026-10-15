package runtime

import "testing"

func TestRemovedProviderPolicy(t *testing.T) {
	for _, provider := range []string{"aliyun", "DeepSeek", "weknoracloud", "zhipu", "baidu", "tencent"} {
		if !blockedModelProvider(provider) {
			t.Fatalf("removed provider %q was accepted", provider)
		}
	}
	for _, provider := range []string{"openai", "anthropic", "generic", "ollama", "cohere", "huggingface_tei", "pinecone"} {
		if blockedModelProvider(provider) {
			t.Fatalf("supported provider %q was rejected", provider)
		}
	}
}

func TestRemovedEndpointPolicy(t *testing.T) {
	for _, endpoint := range []string{
		"https://dashscope.aliyuncs.com/v1",
		"https://api.deepseek.com/v1",
		"https://weknora.weixin.qq.com/api/v1",
		"https://lkeap.tencentcloudapi.com",
		"https://qianfan.baidubce.com/v2",
		"https://api.minimax.cn/v1",
		"https://api.minimax.io/anthropic",
	} {
		if !blockedModelEndpoint(endpoint) {
			t.Fatalf("removed endpoint %q was accepted", endpoint)
		}
	}
	for _, endpoint := range []string{
		"http://localhost:11434/v1",
		"https://api.openai.com/v1",
		"https://api.cohere.com/v2",
		"https://api.pinecone.io",
		"https://api.deepseek.com.example.org/v1",
	} {
		if blockedModelEndpoint(endpoint) {
			t.Fatalf("supported endpoint %q was rejected", endpoint)
		}
	}
}
