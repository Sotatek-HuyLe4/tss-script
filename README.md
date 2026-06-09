# tss-script

A Go demonstration and reference implementation of **Threshold Signature Scheme (TSS)** for ECDSA on the secp256k1 curve. The project wraps [bnb-chain/tss-lib/v2](https://github.com/bnb-chain/tss-lib) to run distributed key generation, threshold signing, key resharing, and signature verification entirely in-process using goroutines and channels as a simulated network layer.

This repository is intended as a learning tool and integration prototype for treasury or custody systems that need multi-party ECDSA without ever assembling a full private key on a single machine.

---

## Table of Contents

- [Overview](#overview)
- [What is Threshold ECDSA?](#what-is-threshold-ecdsa)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Demo Walkthrough](#demo-walkthrough)
- [Core API](#core-api)
- [Configuration](#configuration)
- [How It Works](#how-it-works)
- [Message Encoding](#message-encoding)
- [Key Resharing](#key-resharing)
- [Threshold Semantics](#threshold-semantics)
- [Dependencies](#dependencies)
- [Limitations](#limitations)
- [License](#license)

---

## Overview

`tss-script` models a small committee of parties that jointly:

1. **Generate** a shared ECDSA key pair (each party holds a secret share).
2. **Sign** messages when enough parties participate (threshold signing).
3. **Reshare** key material when committee membership changes (remove one party, add another).
4. **Verify** signatures against the group public key.

All protocol rounds are executed locally. Instead of real network sockets, parties communicate by routing `tss.Message` values through Go channels—useful for understanding the protocol flow before wiring up production transport (gRPC, WebSocket, etc.).

Default demo parameters:

| Parameter       | Value | Meaning                                      |
|-----------------|-------|----------------------------------------------|
| `TOTAL_PARTIES` | `3`   | Committee size (n)                           |
| `THRESHOLD`     | `1`   | Fault tolerance; signing needs `t + 1 = 2` parties |

---

## What is Threshold ECDSA?

In a standard ECDSA setup, one private key signs transactions. In **threshold ECDSA (TSS)**, the private key is never stored whole. It is split into `n` shares distributed among `n` parties. Any subset of at least `t` parties (here, `threshold + 1`) can collaborate to produce a valid ECDSA signature, while fewer than `t` parties learn nothing about the key.

This project uses the **GG18** protocol implementation from `tss-lib`, on curve **secp256k1** (`tss.S256()`), which is the same curve used by Bitcoin and Ethereum.

---

## Features

- **Distributed key generation (DKG)** — GG18 ECDSA keygen with per-party pre-parameters.
- **Threshold signing** — Multi-round signing protocol among a quorum of parties.
- **Key resharing** — Rotate committee membership while preserving the same group public key.
- **Signature verification** — Standard `crypto/ecdsa` verification against the derived group public key.
- **In-process message routing** — Simulated P2P with broadcast and point-to-point delivery.
- **Structured demo** — `main.go` runs a full lifecycle: keygen → sign → verify → reshare → sign again.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                             │
│              (orchestrates the demo lifecycle)              │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │  tss.go  │    │tss_party │    │ utils.go │
    │  TSS     │    │ TSSParty │    │ helpers  │
    └────┬─────┘    └──────────┘    └──────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│              github.com/bnb-chain/tss-lib/v2                │
│   keygen │ signing │ resharing │ common │ tss (routing)     │
└─────────────────────────────────────────────────────────────┘
```

Each `TSSParty` owns:

- A `tss.PartyID` (identity in the protocol).
- `keygen.LocalPreParams` (safe primes generated before keygen).
- `keygen.LocalPartySaveData` (the party's key share and group public key after keygen/resharing).

The `TSS` struct holds the committee configuration and a slice of parties. Protocol logic in `tss.go` spawns `LocalParty` instances from `tss-lib`, routes messages between them, and collects results from end channels.

---

## Project Structure

```
tss-script/
├── main.go        # Entry point; runs the 6-step demo
├── tss.go         # TSS coordinator: keygen, signing, resharing
├── tss_party.go   # Per-party state and pre-parameter generation
├── utils.go       # Message encoding, signature verification, resharing router
├── go.mod         # Module definition and dependencies
├── go.sum         # Dependency checksums
└── README.md
```

| File           | Responsibility                                                                 |
|----------------|--------------------------------------------------------------------------------|
| `main.go`      | Constants, demo steps, console output                                          |
| `tss.go`       | `TSS` type, `NewTSS`, `CreateParties`, `GenerateKey`, `SignMessage`, `ReSharingKey` |
| `tss_party.go` | `TSSParty` type, party construction, `GeneratePreParams`                       |
| `utils.go`     | `StringToBigInt`, `VerifySignature`, `routeResharingMessage`                   |

---

## Requirements

- **Go** 1.25.6 or compatible (see `go.mod`)
- Network access for initial `go mod download` (if dependencies are not cached)

---

## Installation

Clone the repository and fetch dependencies:

```bash
git clone <repository-url>
cd tss-script
go mod download
```

---

## Quick Start

Run the demo:

```bash
go run .
```

Expected output includes six labeled steps: create TSS instance, create parties, generate key, sign and verify, reshare key, then sign and verify again. Public key coordinates and signature `(R, S)` values are printed to stdout.

Build a binary:

```bash
go build -o tss-script .
./tss-script
```

---

## Demo Walkthrough

`main.go` executes the following pipeline:

### Step 1 — Create TSS instance

```go
tss, err := NewTSS(TOTAL_PARTIES, THRESHOLD)  // 3 parties, threshold 1
```

Validates that `totalParties > 0`, `threshold > 0`, and `totalParties >= threshold`.

### Step 2 — Create parties

```go
tss.CreateParties()
```

- Instantiates `TOTAL_PARTIES` `TSSParty` values with IDs `1..n`.
- Sorts party IDs per `tss-lib` conventions.
- Generates `LocalPreParams` in parallel (safe primes; up to 2 minutes per party).

### Step 3 — Generate key

```go
tss.GenerateKey()
```

Runs the GG18 distributed key generation protocol. After completion, each party holds a `KeyShare`, and all shares correspond to the same group ECDSA public key (`KeyShare.ECDSAPub`). The demo prints `X` and `Y` coordinates of the public key.

### Step 4 — Sign message

```go
signatureData, err := tss.SignMessage("Hello, world!")
```

Signs the message using the first `threshold + 1` parties (2 of 3 by default). Each party's key share is used to verify the signature independently:

```go
VerifySignature("Hello, world!", signatureData, party.KeyShare)
```

### Step 5 — Re-share key

```go
tss.ReSharingKey(3, 4)
```

Removes party **3** from the committee and adds a new party **4**. Remaining parties receive updated key shares and new `PartyID` key integers. The group public key remains unchanged (verified by printing the same `X`, `Y` after resharing).

### Step 6 — Re-sign message

Signs `"Hello, world!"` again with the updated committee and verifies the signature for every party.

---

## Core API

### `TSS`

| Method            | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `NewTSS(n, t)`    | Create a TSS session with `n` parties and threshold `t`.                    |
| `CreateParties()` | Allocate parties, sort IDs, generate pre-params.                            |
| `GenerateKey()`   | Run distributed key generation; populate each party's `KeyShare`.         |
| `SignMessage(msg)`| Run threshold signing; returns `*common.SignatureData` with `R` and `S`.   |
| `ReSharingKey(oldId, newId)` | Reshare keys: drop `oldId`, add `newId`.                         |
| `Print()`         | Debug print of configuration and parties.                                   |

### `TSSParty`

| Field / Method       | Description                                              |
|----------------------|----------------------------------------------------------|
| `Id`                 | Human-readable party index (1-based in `NewTSSParty`).   |
| `PartyID`            | `tss-lib` party identity (moniker, id string, key int).  |
| `PreParams`          | Pre-generated safe primes for keygen.                    |
| `KeyShare`           | Post-keygen (or post-resharing) local save data.         |
| `GeneratePreParams()`| Generate pre-params for a single party (used in resharing). |

### Utilities (`utils.go`)

| Function            | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `StringToBigInt`    | UTF-8 string → hex → `*big.Int` (signing payload).                        |
| `VerifySignature`   | Verify `(R, S)` against message and a party's `KeyShare.ECDSAPub`.          |
| `routeResharingMessage` | Internal router for resharing protocol messages (mirrors `tss-lib` tests). |

---

## Configuration

Edit constants in `main.go`:

```go
const (
    TOTAL_PARTIES = 3  // committee size (n)
    THRESHOLD     = 1  // tss-lib threshold (signing quorum = THRESHOLD + 1)
)
```

When integrating into a larger system, pass these values to `NewTSS` instead of hard-coding them.

**Pre-parameter timeout:** `keygen.GeneratePreParams(2 * time.Minute)` is used in `tss.go` and `tss_party.go`. Safe prime generation can be slow; increase the timeout on constrained hardware if keygen fails.

---

## How It Works

### Simulated networking

`tss-lib` parties emit messages on `outCh` channels. This project runs goroutines that:

1. Read outgoing messages from each party.
2. Inspect routing metadata (`IsBroadcast`, `routing.To`).
3. Deliver wire bytes to peer parties via `UpdateFromBytes` (keygen/signing) or `Update` (resharing).

This pattern matches how you would integrate real transport: serialize `tss.Message`, send to peers, deserialize, and call `Update` on the receiving `LocalParty`.

### Key generation flow

```
Party 1 ──┐
Party 2 ──┼──► [message router goroutines] ──► endCh ──► KeyShare per party
Party 3 ──┘
```

1. Build sorted `PartyIDs` and `PeerContext`.
2. Create one `keygen.LocalParty` per committee member.
3. Start all parties concurrently.
4. Route messages until each party writes to its `endCh`.
5. Store `LocalPartySaveData` on each `TSSParty`.

### Signing flow

Uses the first `threshold + 1` parties from `t.Parties`. Each selected party runs `signing.NewLocalParty` with the message hash (`*big.Int`) and its key share. One `SignatureData` is returned from the first end channel; remaining end channels are drained.

### Resharing flow

`ReSharingKey` builds **old** and **new** committees:

- **Old committee:** all current parties (including the one being removed).
- **New committee:** all parties except `oldId`, plus a freshly created party `newId`.

New committee members (except the brand-new party) get regenerated `PartyID` key integers (`party.Id + newId`) so identities are distinct for the resharing round. A single shared `outCh` feeds `routeResharingMessage`, which dispatches to old and/or new `resharing.LocalParty` instances based on message destination flags (`IsToOldCommittee`, `IsToOldAndNewCommittees`).

---

## Message Encoding

`StringToBigInt` converts a string to the signing payload:

1. Encode the string as UTF-8 bytes.
2. Hex-encode those bytes.
3. Parse the hex string as a big-endian integer.

Example: `"Hello, world!"` → hex of raw bytes → `*big.Int`.

`VerifySignature` uses `msgBigInt.Bytes()` as the digest passed to `ecdsa.Verify`. Callers must use the same encoding for signing and verification.

---

## Key Resharing

`ReSharingKey(oldId, newId int)` supports committee rotation:

| Argument | Role                                                                 |
|----------|----------------------------------------------------------------------|
| `oldId`  | Party ID to remove from the new committee after resharing completes. |
| `newId`  | Party ID for the newly joined member.                                |

After a successful resharing:

- `t.Parties` is replaced with the new committee.
- Each surviving party holds an updated `KeyShare`.
- The group public key (`ECDSAPub`) is unchanged—signatures remain valid under the same address/key.

The demo calls `ReSharingKey(3, 4)`: party 3 leaves, party 4 joins, parties 1 and 2 are refreshed with new shares.

---

## Threshold Semantics

In `tss-lib`, the `threshold` parameter to `tss.NewParameters` is **not** the signing quorum directly. The signing quorum is **`threshold + 1`**.

| `THRESHOLD` | Parties required to sign | With `TOTAL_PARTIES = 3` |
|-------------|--------------------------|---------------------------|
| `1`         | `2`                      | 2-of-3                    |
| `2`         | `3`                      | 3-of-3                    |

`SignMessage` explicitly uses `threshold := t.Threshold + 1` and participates with the first `threshold` parties in the sorted slice.

---

## Dependencies

| Package                         | Version  | Purpose                              |
|---------------------------------|----------|--------------------------------------|
| `github.com/bnb-chain/tss-lib/v2` | v2.0.2 | GG18 ECDSA keygen, sign, reshare   |

Indirect dependencies include `btcsuite/btcd` (secp256k1), `golang.org/x/crypto`, and logging/protobuf stacks pulled in by `tss-lib`.

The module path is `github.com/Sotatek-HuyLe4/tss-script` (see `go.mod`).

---

## Limitations

This repository is a **local simulation**, not production-ready custody software:

- **No persistence** — Key shares and pre-params exist only in memory.
- **No authentication or encryption** — Messages are passed in-process without TLS or party authentication.
- **No fault handling for production** — Errors in pre-param generation call `panic`; resharing errors on `errCh` also panic.
- **Fixed signing subset** — `SignMessage` always uses the first `threshold + 1` parties, not an arbitrary quorum selection.
- **No tests** — Behavior is validated only through the `main` demo and manual inspection of output.
- **Pre-param cost** — Initial keygen can take noticeable time while safe primes are generated.

For production treasury systems, add secure storage, authenticated transport, robust error handling, party discovery, and comprehensive tests before handling real assets.

---

## License

Refer to the repository's license file or your organization's policy. If no license is present, contact the repository owner for usage terms.
