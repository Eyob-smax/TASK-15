package domain

import "github.com/google/uuid"

// SystemActorID is the reserved actor ID for all system-initiated operations
// (scheduled jobs, automated processes). It corresponds to the inactive system
// user seeded in seed.sql. The login flow explicitly rejects inactive accounts,
// so this ID can never be used for interactive login.
var SystemActorID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
