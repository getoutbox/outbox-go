# Required Server-Side Changes

Changes documented here are required for correctness or consistency but depend on
server-side proto / API changes. The SDK is implemented as if these changes are
already in place.

## ListChannelsResponse.TotalSize: int32 → int64

**Depends on:** `channels.go` `List` method
**Current state:** `ListChannelsResponse.TotalSize` is `int32` in the proto, while all
other `List*Response` messages use `int64`. The SDK casts `int32` to `int64` at
`channels.go:47`.
**Required change:** Update the `ListChannelsResponse` proto message to use `int64`
for `TotalSize`, matching `ListConnectorsResponse`, `ListAccountsResponse`,
`ListDestinationsResponse`, and `ListMessagesResponse`.
**Impact:** Eliminates the explicit cast and ensures no overflow for result sets
larger than 2^31−1.
