package client

import (
	"time"
)

// DefaultCacheTTL is the default cache TTL for compliance metadata.
// Set to 24 hours as compliance metadata changes infrequently.
// A zero value means no expiration, but we use a long TTL to balance
// cache efficiency with eventual consistency.
const DefaultCacheTTL = 24 * time.Hour

// DefaultCacheCapacity is the default fixed capacity for Otter cache in number of entries.
// Set to 100000 entries to provide reasonable capacity for compliance metadata caching.
const DefaultCacheCapacity = 100000

// CacheKeySeparator is the separator used to create composite cache keys
// from policy engine name and policy rule id.
const CacheKeySeparator = ":"
