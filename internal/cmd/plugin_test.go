package cmd

import "testing"

func TestPluginResponseMeta(t *testing.T) {
	withURL := pluginResponseMeta("amazon/pdp", "https://example.com")
	if withURL["plugin_path"] != "amazon/pdp" {
		t.Fatalf("meta=%v", withURL)
	}
	if withURL["url"] != "https://example.com" {
		t.Fatalf("meta=%v", withURL)
	}

	withoutURL := pluginResponseMeta("amazon/pdp", "   ")
	if withoutURL["plugin_path"] != "amazon/pdp" {
		t.Fatalf("meta=%v", withoutURL)
	}
	if _, ok := withoutURL["url"]; ok {
		t.Fatalf("expected url to be omitted: %v", withoutURL)
	}
}
