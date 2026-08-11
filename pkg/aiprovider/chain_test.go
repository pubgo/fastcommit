package aiprovider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainAvailableRequiresRealProvider(t *testing.T) {
	unavailable := &OpenAIProvider{}
	chain := NewChain(unavailable)
	require.False(t, chain.Available())

	chain = NewChain(unavailable, NewRuleFallback())
	require.True(t, chain.Available())
}
